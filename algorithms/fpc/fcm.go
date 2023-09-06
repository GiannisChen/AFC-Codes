package fpc

type FcmPredictor struct {
	table    []uint64
	lastHash uint64
	mask     uint64
}

func NewFcmPredictor(tableSize uint64) *FcmPredictor {
	return &FcmPredictor{
		table:    make([]uint64, tableSize),
		lastHash: 0,
		mask:     tableSize - 1,
	}
}

func (fp *FcmPredictor) PredictNext() uint64 {
	return fp.table[fp.lastHash]
}

func (fp *FcmPredictor) Update(value uint64) {
	fp.table[fp.lastHash] = value
	fp.lastHash = ((fp.lastHash << 6) ^ (value >> 48)) & fp.mask
}
