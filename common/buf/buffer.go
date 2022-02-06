package buf

import (
	"io"

	"github.com/vmessocket/vmessocket/common/bytespool"
)

const Size = 2048

var pool = bytespool.GetPool(Size)

type Buffer struct {
	v     []byte
	start int32
	end   int32
}

func New() *Buffer {
	return &Buffer{
		v: pool.Get().([]byte),
	}
}

func StackNew() Buffer {
	return Buffer{
		v: pool.Get().([]byte),
	}
}

func (b *Buffer) Release() {
	if b == nil || b.v == nil {
		return
	}

	p := b.v
	b.v = nil
	b.Clear()
	pool.Put(p)
}

func (b *Buffer) Advance(from int32) {
	if from < 0 {
		from += b.Len()
	}
	b.start += from
}

func (b *Buffer) Byte(index int32) byte {
	return b.v[b.start+index]
}

func (b *Buffer) Bytes() []byte {
	return b.v[b.start:b.end]
}

func (b *Buffer) BytesFrom(from int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	return b.v[b.start+from : b.end]
}

func (b *Buffer) BytesRange(from, to int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start+from : b.start+to]
}

func (b *Buffer) BytesTo(to int32) []byte {
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start : b.start+to]
}

func (b *Buffer) Clear() {
	b.start = 0
	b.end = 0
}

func (b *Buffer) Extend(n int32) []byte {
	end := b.end + n
	if end > int32(len(b.v)) {
		panic("extending out of bound")
	}
	ext := b.v[b.end:end]
	b.end = end
	return ext
}

func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

func (b *Buffer) IsFull() bool {
	return b != nil && b.end == int32(len(b.v))
}

func (b *Buffer) Len() int32 {
	if b == nil {
		return 0
	}
	return b.end - b.start
}

func (b *Buffer) Read(data []byte) (int, error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	nBytes := copy(data, b.v[b.start:b.end])
	if int32(nBytes) == b.Len() {
		b.Clear()
	} else {
		b.start += int32(nBytes)
	}
	return nBytes, nil
}

func (b *Buffer) ReadFrom(reader io.Reader) (int64, error) {
	n, err := reader.Read(b.v[b.end:])
	b.end += int32(n)
	return int64(n), err
}

func (b *Buffer) ReadFullFrom(reader io.Reader, size int32) (int64, error) {
	end := b.end + size
	if end > int32(len(b.v)) {
		v := end
		return 0, newError("out of bound: ", v)
	}
	n, err := io.ReadFull(reader, b.v[b.end:end])
	b.end += int32(n)
	return int64(n), err
}

func (b *Buffer) Resize(from, to int32) {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	if to < from {
		panic("Invalid slice")
	}
	b.end = b.start + to
	b.start += from
}

func (b *Buffer) SetByte(index int32, value byte) {
	b.v[b.start+index] = value
}

func (b *Buffer) String() string {
	return string(b.Bytes())
}

func (b *Buffer) Write(data []byte) (int, error) {
	nBytes := copy(b.v[b.end:], data)
	b.end += int32(nBytes)
	return nBytes, nil
}

func (b *Buffer) WriteByte(v byte) error {
	if b.IsFull() {
		return newError("buffer full")
	}
	b.v[b.end] = v
	b.end++
	return nil
}

func (b *Buffer) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}
