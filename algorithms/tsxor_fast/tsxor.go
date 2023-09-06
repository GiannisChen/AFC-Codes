package tsxor_fast

import (
	"afc/common"
	"math/bits"
)

type dictionaryDelta struct {
	size     int
	maxZeros int
	maxIndex int
	tmpZeros int
	isXor    bool
	left     int
	right    int
}

func (d *dictionaryDelta) Add() {
	if d.size < 7 {
		d.size++
		d.right++
	} else {
		d.left++
		d.right++
	}
}

func (d *dictionaryDelta) Search(nums []uint64, num uint64) (uint8, bool, uint8) {
	d.maxZeros = 1
	d.maxIndex = 0xff
	for i := d.right - 1; i >= d.left; i-- {
		if num == nums[i] {
			return 0, false, uint8(i - d.left)
		}
		d.tmpZeros = bits.LeadingZeros64(num^nums[i])/8 + bits.TrailingZeros64(num^nums[i])/8
		if d.tmpZeros > d.maxZeros {
			d.maxZeros = d.tmpZeros
			d.maxIndex = d.right - i - 1
			d.isXor = true
		}
		d.tmpZeros = bits.LeadingZeros64(num-nums[i])/8 + bits.TrailingZeros64(num-nums[i])/8
		if d.tmpZeros > d.maxZeros {
			d.maxZeros = d.tmpZeros
			d.maxIndex = d.right - i - 1
			d.isXor = false
		}
	}
	return 1, d.isXor, uint8(d.maxIndex)
}

func (d *dictionaryDelta) Calculate(nums []uint64, isXor bool, offset uint8, v uint64) (tz int, length int, xor uint64) {
	length = 8 - d.maxZeros
	if isXor {
		xor = nums[d.right-int(offset)-1] ^ v
	} else {
		xor = v - nums[d.right-int(offset)-1]
	}
	tz = bits.TrailingZeros64(xor) / 8
	xor >>= tz * 8
	return
}

func (d *dictionaryDelta) At(offset uint8) int {
	return d.right - int(offset) - 1
}

const (
	small52 = 0x000f_ffff_ffff_ffff
	big12   = 0xfff0_0000_0000_0000
)

func (d *dictionaryDelta) Predict(nums []uint64) uint64 {
	pred := (nums[d.right-1] & nums[d.right-2]) | (nums[d.right-2] & nums[d.right-3]) | (nums[d.right-3]&nums[d.right-1])&small52
	if (nums[d.right-2]&big12)<<1 == (nums[d.right-1]&big12)+(nums[d.right-3]&big12) {
		return (nums[d.right-1]&big12)<<1 - (nums[d.right-2] & big12)
	}
	return pred | (nums[d.right-1] & big12)
}

func Compress(dst []byte, src []uint64) []byte {
	d := dictionaryDelta{}
	dst = common.Append64(dst, src[0])
	d.Add()
	for _, num := range src[1:] {
		v := num
		if t, isXor, offset := d.Search(src, num); t == 0 {
			dst = append(dst, offset)
		} else {
			if offset == 0xff {
				dst = append(dst, 0xff)
				dst = common.Append64(dst, v)
			} else {
				dst = append(dst, 0b1000_0000|offset)
				tz, length, xor := d.Calculate(src, isXor, offset, v)
				if isXor {
					dst = append(dst, uint8(length)|(uint8(tz)<<4))
				} else {
					dst = append(dst, uint8(length)|(uint8(tz)<<4)|0b1000_0000)
				}
				for length > 0 {
					dst = append(dst, uint8(xor))
					xor >>= 8
					length--
				}
			}
		}
		d.Add()
	}
	return dst
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	d := dictionaryDelta{}
	v, i, err := common.Get64(src, 0)
	if err != nil {
		return nil, err
	}
	dst = append(dst, v)
	d.Add()
	i++

	for ; i < len(src); i++ {
		if src[i] == 0xff {
			v, i, err = common.Get64(src, i+1)
			if err != nil {
				return nil, err
			}
		} else {
			if src[i]&0b1000_0000 == 0 {
				v = uint64(dst[d.At(src[i])])
			} else {
				offset := src[i] & 0b0111_1111
				i++
				tz := int((src[i]>>4)&0b0111) * 8
				isXor := (src[i] & 0b1000_0000) == 0
				length := src[i] & 0x0f
				v = uint64(0)
				for j := 0; j < int(length); j++ {
					i++
					v |= uint64(src[i]) << (j * 8)
				}
				v <<= tz
				if isXor {
					v ^= dst[d.At(offset)]
				} else {
					v += dst[d.At(offset)]
				}
			}
		}

		dst = append(dst, v)
		d.Add()
	}
	return dst, nil
}
