package lz4

import (
	"encoding/binary"
	"github.com/bkaradzic/go-lz4"
)

func Compress(dst []byte, src []uint64) []byte {
	uncb := make([]byte, len(src)*8)
	for i, u := range src {
		binary.LittleEndian.PutUint64(uncb[i*8:(i+1)*8], u)
	}
	dst, _ = lz4.Encode(dst, uncb)
	return dst
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	var uncb []byte
	uncb, _ = lz4.Decode(uncb, src)
	for i := 0; i < len(uncb)/8; i++ {
		dst = append(dst, binary.LittleEndian.Uint64(uncb[i*8:(i+1)*8]))
	}
	return dst, nil
}
