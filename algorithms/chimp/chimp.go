package chimp

import (
	"afc/common"
	"math/bits"
)

func Compress(dst []byte, src []uint64) []byte {
	bs := &common.ByteWrapper{Stream: &dst, Count: 0}
	bs.AppendBits(uint64(len(src)), 14)
	bs.AppendBits(src[0], 64)
	lastLeading := 0
	lastValue := src[0]
	for i := 1; i < len(src); i++ {
		xor := src[i] ^ lastValue
		leading := bits.LeadingZeros64(xor)
		if leading > 15 {
			leading = 15
		}
		if leading%2 == 1 {
			leading -= 1
		}
		trailing := bits.TrailingZeros64(xor)
		if trailing > 6 {
			bs.AppendBit(common.Zero)
			if xor == 0 {
				bs.AppendBit(common.Zero)
			} else {
				bs.AppendBit(common.One)
				bs.AppendBits(uint64(leading/2), 3)
				centerBitsCounts := 64 - leading - trailing
				bs.AppendBits(uint64(centerBitsCounts), 6)
				bs.AppendBits(uint64(xor>>trailing), centerBitsCounts)
			}
		} else {
			bs.AppendBit(common.One)
			if lastLeading == leading {
				bs.AppendBit(common.Zero)
				bs.AppendBits(xor, 64-leading)
			} else {
				bs.AppendBit(common.One)
				bs.AppendBits(uint64(leading/2), 3)
				bs.AppendBits(xor, 64-leading)
			}
		}
		lastLeading = leading
		lastValue = src[i]
	}
	return dst
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	bs := &common.ByteWrapper{Stream: &src, Count: 8}
	length, err := bs.ReadBits(14)
	if err != nil {
		return nil, err
	}
	firstValue, err := bs.ReadBits(64)
	if err != nil {
		return nil, err
	}
	dst = append(dst, firstValue)
	lastLeading := 0
	lastValue := firstValue
	curValue := uint64(0)
	for i := uint64(1); i < length; i++ {
		bit, err := bs.ReadBit()
		if err != nil {
			return nil, err
		}
		if !bit {
			bit, err = bs.ReadBit()
			if err != nil {
				return nil, err
			}
			if !bit {
				curValue = lastValue
			} else {
				leading, err := bs.ReadBits(3)
				if err != nil {
					return nil, err
				}
				leading *= 2
				centerCount, err := bs.ReadBits(6)
				if err != nil {
					return nil, err
				}
				xored, err := bs.ReadBits(int(centerCount))
				xored <<= 64 - leading - centerCount
				curValue = lastValue ^ xored
			}
		} else {
			bit, err = bs.ReadBit()
			if err != nil {
				return nil, err
			}
			if !bit {
				xored, err := bs.ReadBits(64 - lastLeading)
				if err != nil {
					return nil, err
				}
				curValue = xored ^ lastValue
			} else {
				leading, err := bs.ReadBits(3)
				if err != nil {
					return nil, err
				}
				leading *= 2
				xored, err := bs.ReadBits(64 - int(leading))
				curValue = xored ^ lastValue
			}
		}
		dst = append(dst, curValue)
		leading := bits.LeadingZeros64(curValue ^ lastValue)
		if leading > 15 {
			leading = 15
		}
		if leading%2 == 1 {
			leading -= 1
		}
		lastLeading = leading
		lastValue = curValue
	}

	return dst, nil
}
