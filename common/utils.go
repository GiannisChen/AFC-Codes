package common

import "io"

func Append64(dst []byte, src uint64) []byte {
	dst = append(dst, uint8(src>>56))
	dst = append(dst, uint8(src>>48))
	dst = append(dst, uint8(src>>40))
	dst = append(dst, uint8(src>>32))
	dst = append(dst, uint8(src>>24))
	dst = append(dst, uint8(src>>16))
	dst = append(dst, uint8(src>>8))
	dst = append(dst, uint8(src))
	return dst
}

func Get64(src []byte, i int) (uint64, int, error) {
	if src == nil || len(src[i:]) < 8 {
		return 0, i, io.EOF
	}
	v := uint64(0)
	for j := 0; j < 8; j++ {
		v <<= 8
		v |= uint64(src[i])
		i++
	}
	return v, i - 1, nil
}
