package mux

import (
	"io"

	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/crypto"
	"github.com/vmessocket/vmessocket/common/serial"
)

type PacketReader struct {
	reader io.Reader
	eof    bool
}

func NewPacketReader(reader io.Reader) *PacketReader {
	return &PacketReader{
		reader: reader,
		eof:    false,
	}
}

func (r *PacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if r.eof {
		return nil, io.EOF
	}

	size, err := serial.ReadUint16(r.reader)
	if err != nil {
		return nil, err
	}

	if size > buf.Size {
		return nil, newError("packet size too large: ", size)
	}

	b := buf.New()
	if _, err := b.ReadFullFrom(r.reader, int32(size)); err != nil {
		b.Release()
		return nil, err
	}
	r.eof = true
	return buf.MultiBuffer{b}, nil
}

func NewStreamReader(reader *buf.BufferedReader) buf.Reader {
	return crypto.NewChunkStreamReaderWithChunkCount(crypto.PlainChunkSizeParser{}, reader, 1)
}
