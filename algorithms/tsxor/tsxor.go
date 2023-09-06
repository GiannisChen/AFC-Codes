package tsxor

import (
	"afc/common"
	"math/bits"
)

type dictionary struct {
	window   []uint64
	size     int
	maxZeros int
	maxIndex int
	tmpZeros int
}

// Add insert value from rear
func (d *dictionary) Add(num uint64) {
	d.window = append(d.window, num)
	if d.size == 127 {
		d.window = d.window[1:]
	} else {
		d.size++
	}
}

// Search checks value from rear to start
// first return enum{0: reference, 1: xor & exception} 1 depends on the second return equals to 0xff or not
func (d *dictionary) Search(num uint64) (uint8, uint8) {
	d.maxZeros = 1
	d.maxIndex = 0xff
	for i := d.size - 1; i >= 0; i-- {
		if num == d.window[i] {
			return 0, uint8(i)
		}
		d.tmpZeros = bits.LeadingZeros64(num^d.window[i])/8 + bits.TrailingZeros64(num^d.window[i])/8
		if d.tmpZeros > d.maxZeros {
			d.maxZeros = d.tmpZeros
			d.maxIndex = i
		}
	}
	return 1, uint8(d.maxIndex)
}

func (d *dictionary) Calculate(offset uint8, num uint64) (tz int, length int, xor uint64) {
	length = 8 - d.maxZeros
	xor = d.window[offset] ^ num
	tz = bits.TrailingZeros64(xor) / 8
	xor >>= tz * 8
	return
}

func (d *dictionary) At(offset uint8) uint64 {
	return d.window[offset]
}

func Compress(dst []byte, src []uint64) []byte {
	d := dictionary{window: []uint64{}}
	dst = common.Append64(dst, src[0])
	d.Add(src[0])
	for _, num := range src[1:] {
		v := num
		if t, offset := d.Search(v); t == 0 {
			dst = append(dst, offset)

		} else {
			if offset == 0xff {
				dst = append(dst, 0xff)
				dst = common.Append64(dst, v)
			} else {
				dst = append(dst, 0b1000_0000|offset)
				tz, length, xor := d.Calculate(offset, v)
				dst = append(dst, uint8(length)|(uint8(tz)<<4))
				for length > 0 {
					dst = append(dst, uint8(xor))
					xor >>= 8
					length--
				}
			}
		}
		d.Add(v)
	}
	return dst
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	d := dictionary{window: []uint64{}}
	v, i, err := common.Get64(src, 0)
	if err != nil {
		return nil, err
	}
	dst = append(dst, v)
	d.Add(v)
	i++

	for ; i < len(src); i++ {
		if src[i] == 0xff {
			v, i, err = common.Get64(src, i+1)
			if err != nil {
				return nil, err
			}
		} else {
			if src[i]&0b1000_0000 == 0 {
				v = d.At(src[i])
			} else {
				offset := src[i] & 0b0111_1111
				i++
				tz := int((src[i]>>4)&0x0f) * 8
				length := src[i] & 0x0f
				v = uint64(0)
				for j := 0; j < int(length); j++ {
					i++
					v |= uint64(src[i]) << (j * 8)
				}
				v <<= tz
				v ^= d.At(offset)
			}
		}
		dst = append(dst, v)
		d.Add(v)
	}
	return dst, nil
}
