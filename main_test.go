package main

import (
	afc "afc/algorithms/afc"
	"afc/algorithms/chimp"
	"afc/algorithms/fpc"
	"afc/algorithms/gorillaz"
	"afc/algorithms/lz4"
	"afc/algorithms/lzw"
	"afc/algorithms/snappy"
	"afc/algorithms/tsxor"
	"afc/algorithms/tsxor_fast"
	"math"
	"math/rand"
	"testing"
)

var testcases = []struct {
	algo       string
	compress   func([]byte, []uint64) []byte
	decompress func([]uint64, []byte) ([]uint64, error)
}{
	{"afc", afc.Compress, afc.Decompress},
	{"chimp", chimp.Compress, chimp.Decompress},
	{"fpc", fpc.Compress, fpc.Decompress},
	{"gorillaz", gorillaz.Compress, gorillaz.Decompress},
	{"lz4", lz4.Compress, lz4.Decompress},
	{"lzw", lzw.Compress, lzw.Decompress},
	{"snappy", snappy.Compress, snappy.Decompress},
	{"tsxor", tsxor.Compress, tsxor.Decompress},
	{"tsxorfast", tsxor_fast.Compress, tsxor_fast.Decompress},
}

func TestMockedFloats(t *testing.T) {
	for _, tcase := range testcases {
		t.Run(tcase.algo, func(t *testing.T) {
			testMockedFloats(t, tcase.compress, tcase.decompress)
		})
	}
}

func TestRandFloats(t *testing.T) {
	for _, tcase := range testcases {
		t.Run(tcase.algo, func(t *testing.T) {
			testRandFloats(t, tcase.compress, tcase.decompress)
		})
	}
}

func testMockedFloats(t *testing.T, compress func([]byte, []uint64) []byte, decompress func([]uint64, []byte) ([]uint64, error)) {
	t.Helper()
	int64s := []uint64{1, 2, 3}
	var compressedByte []byte
	compressedByte = compress(compressedByte, int64s)
	var decompressedInt64s []uint64
	decompressedInt64s, err := decompress(decompressedInt64s, compressedByte)
	if err != nil {
		t.Error(err)
	}
	if len(decompressedInt64s) != len(int64s) {
		t.Error("de-compress error")
	}
	for i := 0; i < len(decompressedInt64s); i++ {
		if decompressedInt64s[i] != int64s[i] {
			t.Error("de-compress error")
		}
	}
}

func testRandFloats(t *testing.T, compress func([]byte, []uint64) []byte, decompress func([]uint64, []byte) ([]uint64, error)) {
	t.Helper()
	var float64s []uint64
	rand.Seed(114514)
	for i := 0; i < 8000; i++ {
		float64s = append(float64s, math.Float64bits(rand.Float64()))
	}
	var compressedByte []byte
	compressedByte = compress(compressedByte, float64s)
	var decompressedInt64s []uint64
	decompressedInt64s, err := decompress(decompressedInt64s, compressedByte)
	if err != nil {
		t.Error(err)
	}
	if len(decompressedInt64s) != len(float64s) {
		t.Error("de-compress error")
	}
	for i := 0; i < len(decompressedInt64s); i++ {
		if decompressedInt64s[i] != float64s[i] {
			t.Errorf("de-compress error %d, want %d get %d", i, float64s[i], decompressedInt64s[i])
		}
	}
}
