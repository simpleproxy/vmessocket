package bitmask

func (b byte) Has(bb byte) bool {
	return (b & bb) != 0
}

func (b *byte) Set(bb byte) {
	*b |= bb
}

func (b *byte) Clear(bb byte) {
	*b &= ^bb
}

func (b *byte) Toggle(bb byte) {
	*b ^= bb
}
