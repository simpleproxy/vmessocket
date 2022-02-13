package mux

import (
	"io"

	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/crypto"
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

func NewStreamReader(reader *buf.BufferedReader) buf.Reader {
	return crypto.NewChunkStreamReaderWithChunkCount(crypto.PlainChunkSizeParser{}, reader, 1)
}
