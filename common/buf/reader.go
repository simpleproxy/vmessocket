package buf

import (
	"io"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/errors"
)

type BufferedReader struct {
	Reader  Reader
	Buffer  MultiBuffer
	Spliter func(MultiBuffer, []byte) (MultiBuffer, int)
}

type PacketReader struct {
	io.Reader
}

type SingleReader struct {
	io.Reader
}

func ReadBuffer(r io.Reader) (*Buffer, error) {
	b := New()
	n, err := b.ReadFrom(r)
	if n > 0 {
		return b, err
	}
	b.Release()
	return nil, err
}

func readOneUDP(r io.Reader) (*Buffer, error) {
	b := New()
	for i := 0; i < 64; i++ {
		_, err := b.ReadFrom(r)
		if !b.IsEmpty() {
			return b, nil
		}
		if err != nil {
			b.Release()
			return nil, err
		}
	}
	b.Release()
	return nil, newError("Reader returns too many empty payloads.")
}

func (r *BufferedReader) BufferedBytes() int32 {
	return r.Buffer.Len()
}

func (r *BufferedReader) Close() error {
	return common.Close(r.Reader)
}

func (r *BufferedReader) Interrupt() {
	common.Interrupt(r.Reader)
}

func (r *BufferedReader) Read(b []byte) (int, error) {
	spliter := r.Spliter
	if spliter == nil {
		spliter = SplitBytes
	}
	if !r.Buffer.IsEmpty() {
		buffer, nBytes := spliter(r.Buffer, b)
		r.Buffer = buffer
		if r.Buffer.IsEmpty() {
			r.Buffer = nil
		}
		return nBytes, nil
	}
	mb, err := r.Reader.ReadMultiBuffer()
	if err != nil {
		return 0, err
	}
	mb, nBytes := spliter(mb, b)
	if !mb.IsEmpty() {
		r.Buffer = mb
	}
	return nBytes, nil
}

func (r *BufferedReader) ReadAtMost(size int32) (MultiBuffer, error) {
	if r.Buffer.IsEmpty() {
		mb, err := r.Reader.ReadMultiBuffer()
		if mb.IsEmpty() && err != nil {
			return nil, err
		}
		r.Buffer = mb
	}
	rb, mb := SplitSize(r.Buffer, size)
	r.Buffer = rb
	if r.Buffer.IsEmpty() {
		r.Buffer = nil
	}
	return mb, nil
}

func (r *BufferedReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := r.Read(b[:])
	return b[0], err
}

func (r *BufferedReader) ReadMultiBuffer() (MultiBuffer, error) {
	if !r.Buffer.IsEmpty() {
		mb := r.Buffer
		r.Buffer = nil
		return mb, nil
	}
	return r.Reader.ReadMultiBuffer()
}

func (r *PacketReader) ReadMultiBuffer() (MultiBuffer, error) {
	b, err := readOneUDP(r.Reader)
	if err != nil {
		return nil, err
	}
	return MultiBuffer{b}, nil
}

func (r *SingleReader) ReadMultiBuffer() (MultiBuffer, error) {
	b, err := ReadBuffer(r.Reader)
	return MultiBuffer{b}, err
}

func (r *BufferedReader) WriteTo(writer io.Writer) (int64, error) {
	nBytes, err := r.writeToInternal(writer)
	if errors.Cause(err) == io.EOF {
		return nBytes, nil
	}
	return nBytes, err
}

func (r *BufferedReader) writeToInternal(writer io.Writer) (int64, error) {
	mbWriter := NewWriter(writer)
	var sc SizeCounter
	if r.Buffer != nil {
		sc.Size = int64(r.Buffer.Len())
		if err := mbWriter.WriteMultiBuffer(r.Buffer); err != nil {
			return 0, err
		}
		r.Buffer = nil
	}
	err := Copy(r.Reader, mbWriter, CountSize(&sc))
	return sc.Size, err
}
