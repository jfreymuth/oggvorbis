package vorbis

type huffmanCode struct {
	tree  []uint32
	table [256]uint32 // (value<<4)|length, 0 = fallback to tree
}

func (h *huffmanCode) Lookup(r *bitReader) uint32 {
	if r.bitsLeft < 8 {
		r.refill()
	}
	if r.bitsLeft >= 8 {
		entry := h.table[uint8(r.buf)]
		if entry != 0 {
			r.buf >>= entry & 0xf
			r.bitsLeft -= uint(entry & 0xf)
			return entry >> 4
		}
	}
	// fallback: tree walk for codes longer than 8 bits
	i := uint32(0)
	for i&1 == 0 {
		i = h.tree[i+r.Read1()]
	}
	return i >> 1
}

type huffmanBuilder struct {
	tree      []uint32
	minLength []uint8
}

func newHuffmanBuilder(size uint32) *huffmanBuilder {
	return &huffmanBuilder{
		tree:      make([]uint32, size),
		minLength: make([]uint8, size/2),
	}
}

func (t *huffmanBuilder) Put(entry uint32, length uint8) {
	t.put(0, entry, length-1)
}

func (t *huffmanBuilder) put(index, entry uint32, length uint8) bool {
	if length < t.minLength[index/2] {
		return false
	}
	if length == 0 {
		if t.tree[index] == 0 {
			t.tree[index] = entry*2 + 1
			return true
		}
		if t.tree[index+1] == 0 {
			t.tree[index+1] = entry*2 + 1
			t.minLength[index/2] = 1
			return true
		}
		t.minLength[index/2] = 1
		return false
	}
	if t.tree[index]&1 == 0 {
		if t.tree[index] == 0 {
			t.tree[index] = t.findEmpty(index + 2)
		}
		if t.put(t.tree[index], entry, length-1) {
			return true
		}
	}
	if t.tree[index+1]&1 == 0 {
		if t.tree[index+1] == 0 {
			t.tree[index+1] = t.findEmpty(index + 2)
		}
		if t.put(t.tree[index+1], entry, length-1) {
			return true
		}
	}
	t.minLength[index/2] = length + 1
	return false
}

func (t *huffmanBuilder) findEmpty(index uint32) uint32 {
	for t.tree[index] != 0 {
		index += 2
	}
	return index
}

func (t *huffmanBuilder) build() *huffmanCode {
	h := &huffmanCode{tree: t.tree}
	for bits := 0; bits < 256; bits++ {
		i := uint32(0)
		b := uint32(bits)
		for consumed := uint32(1); consumed <= 8; consumed++ {
			next := h.tree[i+b&1]
			b >>= 1
			if next&1 != 0 {
				// leaf: pack (value<<4)|length, guaranteed non-zero since consumed>=1
				h.table[bits] = (next>>1)<<4 | consumed
				break
			}
			if next == 0 {
				break // incomplete tree branch
			}
			i = next
		}
	}
	return h
}
