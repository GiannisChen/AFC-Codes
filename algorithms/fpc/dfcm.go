package fpc

type DfcmPredictor struct {
	table     []uint64
	lastValue uint64
	lastHash  uint64
	mask      uint64
}

func NewDfcmPredictor(tableSize uint64) *DfcmPredictor {
	return &DfcmPredictor{
		table:     make([]uint64, tableSize),
		lastValue: 0,
		lastHash:  0,
		mask:      tableSize - 1,
	}
}

func (fp *DfcmPredictor) PredictNext() uint64 {
	return fp.table[fp.lastHash] + fp.lastValue
}

func (fp *DfcmPredictor) Update(value uint64) {
	fp.table[fp.lastHash] = value - fp.lastValue
	fp.lastHash = ((fp.lastHash << 2) ^ ((value - fp.lastValue) >> 40)) & fp.mask
	fp.lastValue = value
}
