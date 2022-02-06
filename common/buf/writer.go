package buf

import (
	"io"
	"net"
	"sync"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/errors"
)

var (
	Discard      Writer    = noOpWriter(0)
	DiscardBytes io.Writer = noOpWriter(0)
)

type BufferedWriter struct {
	sync.Mutex
	writer   Writer
	buffer   *Buffer
	buffered bool
}

type BufferToBytesWriter struct {
	io.Writer
	cache [][]byte
}

type noOpWriter byte

type SequentialWriter struct {
	io.Writer
}

func NewBufferedWriter(writer Writer) *BufferedWriter {
	return &BufferedWriter{
		writer:   writer,
		buffer:   New(),
		buffered: true,
	}
}

func (w *BufferedWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	return common.Close(w.writer)
}

func (w *BufferedWriter) Flush() error {
	w.Lock()
	defer w.Unlock()
	return w.flushInternal()
}

func (w *BufferedWriter) flushInternal() error {
	if w.buffer.IsEmpty() {
		return nil
	}
	b := w.buffer
	w.buffer = nil
	if writer, ok := w.writer.(io.Writer); ok {
		err := WriteAllBytes(writer, b.Bytes())
		b.Release()
		return err
	}
	return w.writer.WriteMultiBuffer(MultiBuffer{b})
}

func (w *BufferedWriter) ReadFrom(reader io.Reader) (int64, error) {
	if err := w.SetBuffered(false); err != nil {
		return 0, err
	}
	var sc SizeCounter
	err := Copy(NewReader(reader), w, CountSize(&sc))
	return sc.Size, err
}

func (w *BufferToBytesWriter) ReadFrom(reader io.Reader) (int64, error) {
	var sc SizeCounter
	err := Copy(NewReader(reader), w, CountSize(&sc))
	return sc.Size, err
}

func (noOpWriter) ReadFrom(reader io.Reader) (int64, error) {
	b := New()
	defer b.Release()
	totalBytes := int64(0)
	for {
		b.Clear()
		_, err := b.ReadFrom(reader)
		totalBytes += int64(b.Len())
		if err != nil {
			if errors.Cause(err) == io.EOF {
				return totalBytes, nil
			}
			return totalBytes, err
		}
	}
}

func (w *BufferedWriter) SetBuffered(f bool) error {
	w.Lock()
	defer w.Unlock()
	w.buffered = f
	if !f {
		return w.flushInternal()
	}
	return nil
}

func (w *BufferedWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	w.Lock()
	defer w.Unlock()
	if !w.buffered {
		if writer, ok := w.writer.(io.Writer); ok {
			return writer.Write(b)
		}
	}
	totalBytes := 0
	for len(b) > 0 {
		if w.buffer == nil {
			w.buffer = New()
		}
		nBytes, err := w.buffer.Write(b)
		totalBytes += nBytes
		if err != nil {
			return totalBytes, err
		}
		if !w.buffered || w.buffer.IsFull() {
			if err := w.flushInternal(); err != nil {
				return totalBytes, err
			}
		}
		b = b[nBytes:]
	}
	return totalBytes, nil
}

func (noOpWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *BufferedWriter) WriteByte(c byte) error {
	return common.Error2(w.Write([]byte{c}))
}

func (w *BufferedWriter) WriteMultiBuffer(b MultiBuffer) error {
	if b.IsEmpty() {
		return nil
	}
	w.Lock()
	defer w.Unlock()
	if !w.buffered {
		return w.writer.WriteMultiBuffer(b)
	}
	reader := MultiBufferContainer{
		MultiBuffer: b,
	}
	defer reader.Close()
	for !reader.MultiBuffer.IsEmpty() {
		if w.buffer == nil {
			w.buffer = New()
		}
		common.Must2(w.buffer.ReadFrom(&reader))
		if w.buffer.IsFull() {
			if err := w.flushInternal(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *BufferToBytesWriter) WriteMultiBuffer(mb MultiBuffer) error {
	defer ReleaseMulti(mb)
	size := mb.Len()
	if size == 0 {
		return nil
	}
	if len(mb) == 1 {
		return WriteAllBytes(w.Writer, mb[0].Bytes())
	}
	if cap(w.cache) < len(mb) {
		w.cache = make([][]byte, 0, len(mb))
	}
	bs := w.cache
	for _, b := range mb {
		bs = append(bs, b.Bytes())
	}
	defer func() {
		for idx := range bs {
			bs[idx] = nil
		}
	}()
	nb := net.Buffers(bs)
	for size > 0 {
		n, err := nb.WriteTo(w.Writer)
		if err != nil {
			return err
		}
		size -= int32(n)
	}
	return nil
}

func (noOpWriter) WriteMultiBuffer(b MultiBuffer) error {
	ReleaseMulti(b)
	return nil
}

func (w *SequentialWriter) WriteMultiBuffer(mb MultiBuffer) error {
	mb, err := WriteMultiBuffer(w.Writer, mb)
	ReleaseMulti(mb)
	return err
}
