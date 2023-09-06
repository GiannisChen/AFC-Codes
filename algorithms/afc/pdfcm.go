package afc

type PdfcmPredictor struct {
	table         []uint64
	lastHash      uint64
	lastValue     uint64
	lastLeading12 uint64
	mask          uint64
}

func NewPdfcmPredictor(tableSize uint64) *PdfcmPredictor {
	return &PdfcmPredictor{
		table:         make([]uint64, tableSize),
		lastHash:      0,
		lastValue:     0,
		lastLeading12: 0,
		mask:          tableSize - 1,
	}
}

func (fp *PdfcmPredictor) PredictNext() uint64 {
	return (fp.table[fp.lastHash]+fp.lastValue)&0x03FF_FFFF_FFFF_FFFF | fp.lastLeading12
}

func (fp *PdfcmPredictor) Update(value uint64) {
	fp.table[fp.lastHash] = value - fp.lastValue
	fp.lastHash = ((fp.lastHash) ^ ((value - fp.lastValue) >> 50)) & fp.mask
	fp.lastValue = value
	fp.lastLeading12 = value & 0xFC00_0000_0000_0000
}
