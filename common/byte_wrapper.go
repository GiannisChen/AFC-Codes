package common

import "io"

type Bit bool

const (
	Zero Bit = false
	One  Bit = true
)

type ByteWrapper struct {
	Stream *[]byte
	Count  uint8
}

func (bw *ByteWrapper) AppendBit(b Bit) {
	if bw.Count == 0 {
		*bw.Stream = append(*bw.Stream, 0)
		bw.Count = 8
	}
	i := len(*bw.Stream) - 1
	if b {
		(*bw.Stream)[i] |= 1 << (bw.Count - 1)
	}
	bw.Count--
}

func (bw *ByteWrapper) appendByte(byt byte) {

	if bw.Count == 0 {
		*bw.Stream = append(*bw.Stream, 0)
		bw.Count = 8
	}

	i := len(*bw.Stream) - 1

	// fill up bw.bw with bw.count bits from byt
	(*bw.Stream)[i] |= byt >> (8 - bw.Count)

	*bw.Stream = append(*bw.Stream, 0)
	i++
	(*bw.Stream)[i] = byt << bw.Count
}

func (bw *ByteWrapper) AppendBits(u uint64, size int) {
	u <<= 64 - uint(size)
	for size >= 8 {
		byt := byte(u >> 56)
		bw.appendByte(byt)
		u <<= 8
		size -= 8
	}
	for size > 0 {
		bw.AppendBit((u >> 63) == 1)
		u <<= 1
		size--
	}
}

func (bw *ByteWrapper) Finish() {
	// append 11 111111(63) 111111(63) sigbits(63)+leadingZeros(63) > total(64)
	// invalid in encoding.
	bw.appendByte(0xff)
	bw.AppendBits(0x0000_00ff, 6)
	bw.AppendBit(Zero)
}

func (bw *ByteWrapper) ReadBit() (Bit, error) {
	if len(*bw.Stream) == 0 {
		return false, io.EOF
	}

	if bw.Count == 0 {
		*bw.Stream = (*bw.Stream)[1:]
		// did we just run out of stuff to read?
		if len(*bw.Stream) == 0 {
			return false, io.EOF
		}
		bw.Count = 8
	}

	bw.Count--
	d := (*bw.Stream)[0] & 0x80
	(*bw.Stream)[0] <<= 1
	return d != 0, nil
}

func (bw *ByteWrapper) ReadByte() (byte, error) {
	if len(*bw.Stream) == 0 {
		return 0, io.EOF
	}

	if bw.Count == 0 {
		*bw.Stream = (*bw.Stream)[1:]

		if len(*bw.Stream) == 0 {
			return 0, io.EOF
		}

		bw.Count = 8
	}

	if bw.Count == 8 {
		bw.Count = 0
		return (*bw.Stream)[0], nil
	}

	byt := (*bw.Stream)[0]
	*bw.Stream = (*bw.Stream)[1:]

	if len(*bw.Stream) == 0 {
		return 0, io.EOF
	}

	byt |= (*bw.Stream)[0] >> bw.Count
	(*bw.Stream)[0] <<= 8 - bw.Count

	return byt, nil
}

func (bw *ByteWrapper) ReadBits(nbits int) (uint64, error) {
	var u uint64
	for nbits >= 8 {
		byt, err := bw.ReadByte()
		if err != nil {
			return 0, err
		}
		u = (u << 8) | uint64(byt)
		nbits -= 8
	}
	if nbits == 0 {
		return u, nil
	}
	if nbits > int(bw.Count) {
		u = (u << uint(bw.Count)) | uint64((*bw.Stream)[0]>>(8-bw.Count))
		nbits -= int(bw.Count)
		*bw.Stream = (*bw.Stream)[1:]

		if len(*bw.Stream) == 0 {
			return 0, io.EOF
		}
		bw.Count = 8
	}

	u = (u << uint(nbits)) | uint64((*bw.Stream)[0]>>(8-uint(nbits)))
	(*bw.Stream)[0] <<= uint(nbits)
	bw.Count -= uint8(nbits)
	return u, nil
}
