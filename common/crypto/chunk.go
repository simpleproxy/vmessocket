package crypto

import (
	"encoding/binary"
	"io"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
)

type AEADChunkSizeParser struct {
	Auth *AEADAuthenticator
}

type ChunkSizeDecoder interface {
	SizeBytes() int32
	Decode([]byte) (uint16, error)
}

type ChunkSizeEncoder interface {
	SizeBytes() int32
	Encode(uint16, []byte) []byte
}

type ChunkStreamReader struct {
	sizeDecoder ChunkSizeDecoder
	reader      *buf.BufferedReader
	buffer       []byte
	leftOverSize int32
	maxNumChunk  uint32
	numChunk     uint32
}

type ChunkStreamWriter struct {
	sizeEncoder ChunkSizeEncoder
	writer      buf.Writer
}

type PaddingLengthGenerator interface {
	MaxPaddingLen() uint16
	NextPaddingLen() uint16
}

type PlainChunkSizeParser struct{}

func NewChunkStreamReader(sizeDecoder ChunkSizeDecoder, reader io.Reader) *ChunkStreamReader {
	return NewChunkStreamReaderWithChunkCount(sizeDecoder, reader, 0)
}

func NewChunkStreamReaderWithChunkCount(sizeDecoder ChunkSizeDecoder, reader io.Reader, maxNumChunk uint32) *ChunkStreamReader {
	r := &ChunkStreamReader{
		sizeDecoder: sizeDecoder,
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
		maxNumChunk: maxNumChunk,
	}
	if breader, ok := reader.(*buf.BufferedReader); ok {
		r.reader = breader
	} else {
		r.reader = &buf.BufferedReader{Reader: buf.NewReader(reader)}
	}
	return r
}

func NewChunkStreamWriter(sizeEncoder ChunkSizeEncoder, writer io.Writer) *ChunkStreamWriter {
	return &ChunkStreamWriter{
		sizeEncoder: sizeEncoder,
		writer:      buf.NewWriter(writer),
	}
}

func (p *AEADChunkSizeParser) Decode(b []byte) (uint16, error) {
	b, err := p.Auth.Open(b[:0], b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b) + uint16(p.Auth.Overhead()), nil
}

func (PlainChunkSizeParser) Decode(b []byte) (uint16, error) {
	return binary.BigEndian.Uint16(b), nil
}

func (p *AEADChunkSizeParser) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size-uint16(p.Auth.Overhead()))
	b, err := p.Auth.Seal(b[:0], b[:2])
	common.Must(err)
	return b
}

func (PlainChunkSizeParser) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size)
	return b[:2]
}

func (r *ChunkStreamReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	size := r.leftOverSize
	if size == 0 {
		r.numChunk++
		if r.maxNumChunk > 0 && r.numChunk > r.maxNumChunk {
			return nil, io.EOF
		}
		nextSize, err := r.readSize()
		if err != nil {
			return nil, err
		}
		if nextSize == 0 {
			return nil, io.EOF
		}
		size = int32(nextSize)
	}
	r.leftOverSize = size
	mb, err := r.reader.ReadAtMost(size)
	if !mb.IsEmpty() {
		r.leftOverSize -= mb.Len()
		return mb, nil
	}
	return nil, err
}

func (r *ChunkStreamReader) readSize() (uint16, error) {
	if _, err := io.ReadFull(r.reader, r.buffer); err != nil {
		return 0, err
	}
	return r.sizeDecoder.Decode(r.buffer)
}

func (p *AEADChunkSizeParser) SizeBytes() int32 {
	return 2 + int32(p.Auth.Overhead())
}

func (PlainChunkSizeParser) SizeBytes() int32 {
	return 2
}

func (w *ChunkStreamWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	const sliceSize = 8192
	mbLen := mb.Len()
	mb2Write := make(buf.MultiBuffer, 0, mbLen/buf.Size+mbLen/sliceSize+2)
	for {
		mb2, slice := buf.SplitSize(mb, sliceSize)
		mb = mb2
		b := buf.New()
		w.sizeEncoder.Encode(uint16(slice.Len()), b.Extend(w.sizeEncoder.SizeBytes()))
		mb2Write = append(mb2Write, b)
		mb2Write = append(mb2Write, slice...)
		if mb.IsEmpty() {
			break
		}
	}
	return w.writer.WriteMultiBuffer(mb2Write)
}
