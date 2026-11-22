package vorbis

type bitReader struct {
	data     []byte
	position int
	buf      uint64
	bitsLeft uint
	eof      bool
}

func newBitReader(data []byte) *bitReader {
	return &bitReader{data: data}
}

func (r *bitReader) EOF() bool {
	return r.eof
}

func (r *bitReader) refill() {
	for r.bitsLeft <= 56 && r.position < len(r.data) {
		r.buf |= uint64(r.data[r.position]) << r.bitsLeft
		r.bitsLeft += 8
		r.position++
	}
}

func (r *bitReader) Read1() uint32 {
	if r.bitsLeft == 0 {
		r.refill()
		if r.bitsLeft == 0 {
			r.eof = true
			return 0
		}
	}
	bit := uint32(r.buf & 1)
	r.buf >>= 1
	r.bitsLeft--
	return bit
}

func (r *bitReader) read(n uint) uint32 {
	if r.bitsLeft < n {
		r.refill()
		if r.bitsLeft < n {
			r.eof = true
			return 0
		}
	}
	val := uint32(r.buf & ((1 << n) - 1))
	r.buf >>= n
	r.bitsLeft -= n
	return val
}

func (r *bitReader) Read8(n uint) uint8 {
	return uint8(r.read(n))
}

func (r *bitReader) Read16(n uint) uint16 {
	return uint16(r.read(n))
}

func (r *bitReader) Read32(n uint) uint32 {
	return r.read(n)
}

func (r *bitReader) ReadBool() bool {
	return r.Read1() == 1
}
