package lzw

import (
	"bytes"
	"compress/lzw"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

func Compress(dst []byte, src []uint64) []byte {
	uncb := make([]byte, len(src)*8)
	for i, u := range src {
		binary.LittleEndian.PutUint64(uncb[i*8:(i+1)*8], u)
	}
	bw := &bytes.Buffer{}
	lzwEncoder := lzw.NewWriter(bw, lzw.LSB, 8)
	lzwEncoder.Write(uncb)
	lzwEncoder.Close()
	return append(dst, bw.Bytes()...)
}

func Decompress(dst []uint64, src []byte) ([]uint64, error) {
	lzwDecoder := lzw.NewReader(bytes.NewReader(src), lzw.LSB, 8)
	uncb, err := ioutil.ReadAll(lzwDecoder)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(uncb)/8; i++ {
		dst = append(dst, binary.LittleEndian.Uint64(uncb[i*8:(i+1)*8]))
	}
	lzwDecoder.Close()
	return dst, nil
}
