package network

type BitSet struct {
	bytes   []byte
	current int
	bit     int
}

func NewBitSet() *BitSet {
	return &BitSet{
		bytes:   nil,
		current: 0,
		bit:     7,
	}
}

func (b *BitSet) Size() int {
	return len(b.bytes)
}

func (b *BitSet) Write(w *ProtocolWriter) error {
	_, err := w.Write(b.bytes)
	if err != nil {
		return err
	}

	return nil
}

func (b *BitSet) SetBits(v uint32, n int) {
	for i := n - 1; i >= 0; i-- {
		if (v & (1 << i)) != 0 {
			b.Set()
		} else {
			b.Skip()
		}
	}
}

func (b *BitSet) Set() {
	b.set(true)
}

func (b *BitSet) Skip() {
	b.set(false)
}

func (b *BitSet) set(on bool) {
	if len(b.bytes) == 0 || b.current >= len(b.bytes) {
		b.bytes = append(b.bytes, 0x00)
	}

	if on {
		b.bytes[len(b.bytes)-1] |= 1 << b.bit
	}

	b.bit--
	if b.bit < 0 {
		b.current++
		b.bit = 7
	}
}
