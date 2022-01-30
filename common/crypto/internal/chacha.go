package internal

//go:generate go run chacha_core_gen.go

import (
	"encoding/binary"
)

const (
	wordSize  = 4
	stateSize = 16
	blockSize = stateSize * wordSize
)

type ChaCha20Stream struct {
	state  [stateSize]uint32
	block  [blockSize]byte
	offset int
	rounds int
}

func NewChaCha20Stream(key []byte, nonce []byte, rounds int) *ChaCha20Stream {
	s := new(ChaCha20Stream)
	s.state[0] = 0x61707865
	s.state[1] = 0x3320646e
	s.state[2] = 0x79622d32
	s.state[3] = 0x6b206574

	for i := 0; i < 8; i++ {
		s.state[i+4] = binary.LittleEndian.Uint32(key[i*4 : i*4+4])
	}

	switch len(nonce) {
	case 8:
		s.state[14] = binary.LittleEndian.Uint32(nonce[0:])
		s.state[15] = binary.LittleEndian.Uint32(nonce[4:])
	case 12:
		s.state[13] = binary.LittleEndian.Uint32(nonce[0:4])
		s.state[14] = binary.LittleEndian.Uint32(nonce[4:8])
		s.state[15] = binary.LittleEndian.Uint32(nonce[8:12])
	default:
		panic("bad nonce length")
	}

	s.rounds = rounds
	ChaCha20Block(&s.state, s.block[:], s.rounds)
	return s
}

func (s *ChaCha20Stream) XORKeyStream(dst, src []byte) {
	i := 0
	max := len(src)
	for i < max {
		gap := blockSize - s.offset

		limit := i + gap
		if limit > max {
			limit = max
		}

		o := s.offset
		for j := i; j < limit; j++ {
			dst[j] = src[j] ^ s.block[o]
			o++
		}

		i += gap
		s.offset = o

		if o == blockSize {
			s.offset = 0
			s.state[12]++
			ChaCha20Block(&s.state, s.block[:], s.rounds)
		}
	}
}
