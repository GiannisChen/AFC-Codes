package afc

import (
	"afc/common"
	"math/bits"
)

const SIZE int = 1 << LOG

const LOG int = 3

type compressor struct {
	dictSize     int
	dictIndex    int
	dictMaxZeros int
	dictTmpZeros int
	xfcm         *PdfcmPredictor
}

func (c *compressor) update(value uint64) {
	c.xfcm.Update(value)
	if c.dictSize < SIZE {
		c.dictSize++
	}
}

func Compress(dst []byte, src []uint64) []byte {
	bs := &common.ByteWrapper{Stream: &dst, Count: 0}
	c := compressor{xfcm: NewPdfcmPredictor(16)}
	bs.AppendBits(src[0], 64)
	bs.AppendBits(uint64(len(src)), 14)
	c.update(src[0])
	var isReference, isXfcm bool
	for i := 1; i < len(src); i++ {
		if src[i] == src[i-1] {
			bs.AppendBit(common.Zero)
			c.update(src[i])
			continue
		}
		isReference, isXfcm = false, false
		c.dictIndex = -1
		c.dictMaxZeros = 1
		for j := 0; j < c.dictSize; j++ {
			if src[i-j-1] == src[i] {
				bs.AppendBits(0b10<<LOG|uint64(j), 2+LOG)
				isReference = true
				break
			}
			if xor := src[i-j-1] ^ src[i]; bits.LeadingZeros64(xor)/8+bits.TrailingZeros64(xor)/8 > c.dictMaxZeros {
				c.dictIndex = j
				c.dictMaxZeros = bits.LeadingZeros64(xor)/8 + bits.TrailingZeros64(xor)/8
			}
		}
		if !isReference {
			if xor := c.xfcm.PredictNext() ^ src[i]; bits.LeadingZeros64(xor)/8+bits.TrailingZeros64(xor)/8 >= c.dictMaxZeros {
				c.dictMaxZeros = bits.LeadingZeros64(xor)/8 + bits.TrailingZeros64(xor)/8
				isXfcm = true
			}
			if c.dictMaxZeros > 8 {
				c.dictMaxZeros = 8
			}
			if isXfcm {
				bs.AppendBits(0b01110, 4)
				xor := c.xfcm.PredictNext() ^ src[i]
				tz := uint64(bits.TrailingZeros64(xor) / 8)
				bs.AppendBits(tz, 3)
				bs.AppendBits(uint64(8-c.dictMaxZeros), 3)
				bs.AppendBits(xor>>(tz*8), (8-c.dictMaxZeros)*8)
			} else {
				if c.dictIndex == -1 {
					bs.AppendBits(0b01111, 4)
					bs.AppendBits(src[i], 64)
				} else {
					bs.AppendBits(0b0110, 3)
					bs.AppendBits(uint64(c.dictIndex), LOG)
					xor := src[i-c.dictIndex-1] ^ src[i]
					tz := uint64(bits.TrailingZeros64(xor) / 8)
					bs.AppendBits(tz, 3)
					bs.AppendBits(uint64(8-c.dictMaxZeros), 3)
					bs.AppendBits(xor>>(tz*8), (8-c.dictMaxZeros)*8)
				}
			}
		}
		c.update(src[i])
	}
	return dst
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	bs := &common.ByteWrapper{Stream: &src, Count: 8}
	c := compressor{xfcm: NewPdfcmPredictor(16)}
	firstValue, err := bs.ReadBits(64)
	if err != nil {
		return nil, err
	}
	dst = append(dst, firstValue)
	c.update(firstValue)
	length, err := bs.ReadBits(14)
	if err != nil {
		return nil, err
	}
	for i := uint64(1); i < length; i++ {
		bit, err := bs.ReadBit()
		if err != nil {
			return nil, err
		}
		if !bit { // 0'
			dst = append(dst, dst[i-1])
		} else {
			bit, err = bs.ReadBit()
			if err != nil {
				return nil, err
			}
			if !bit { // 10'
				offset, err := bs.ReadBits(LOG)
				if err != nil {
					return nil, err
				}
				dst = append(dst, dst[i-offset-1])
			} else {
				bit, err = bs.ReadBit()
				if err != nil {
					return nil, err
				}
				if !bit { // 110'
					offset, err := bs.ReadBits(LOG)
					if err != nil {
						return nil, err
					}
					tz, err := bs.ReadBits(3)
					if err != nil {
						return nil, err
					}
					l, err := bs.ReadBits(3)
					if err != nil {
						return nil, err
					}
					xor, err := bs.ReadBits(int(l * 8))
					if err != nil {
						return nil, err
					}
					xor <<= tz * 8
					xor ^= dst[i-offset-1]
					dst = append(dst, xor)
				} else {
					bit, err = bs.ReadBit()
					if err != nil {
						return nil, err
					}
					if !bit { //1110'
						if err != nil {
							return nil, err
						}
						tz, err := bs.ReadBits(3)
						if err != nil {
							return nil, err
						}
						l, err := bs.ReadBits(3)
						if err != nil {
							return nil, err
						}
						xor, err := bs.ReadBits(int(l * 8))
						if err != nil {
							return nil, err
						}
						xor <<= tz * 8
						xor ^= c.xfcm.PredictNext()
						dst = append(dst, xor)
					} else { //1111'
						exception, err := bs.ReadBits(64)
						if err != nil {
							return nil, err
						}
						dst = append(dst, exception)
					}
				}
			}
		}
		c.update(dst[i])
	}
	return dst, nil
}
