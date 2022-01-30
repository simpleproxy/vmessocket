package buf

import (
	"io"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/serial"
)

func ReadAllToBytes(reader io.Reader) ([]byte, error) {
	mb, err := ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	if mb.Len() == 0 {
		return nil, nil
	}
	b := make([]byte, mb.Len())
	mb, _ = SplitBytes(mb, b)
	ReleaseMulti(mb)
	return b, nil
}

type MultiBuffer []*Buffer

func MergeMulti(dest MultiBuffer, src MultiBuffer) (MultiBuffer, MultiBuffer) {
	dest = append(dest, src...)
	for idx := range src {
		src[idx] = nil
	}
	return dest, src[:0]
}

func MergeBytes(dest MultiBuffer, src []byte) MultiBuffer {
	n := len(dest)
	if n > 0 && !(dest)[n-1].IsFull() {
		nBytes, _ := (dest)[n-1].Write(src)
		src = src[nBytes:]
	}

	for len(src) > 0 {
		b := New()
		nBytes, _ := b.Write(src)
		src = src[nBytes:]
		dest = append(dest, b)
	}

	return dest
}

func ReleaseMulti(mb MultiBuffer) MultiBuffer {
	for i := range mb {
		mb[i].Release()
		mb[i] = nil
	}
	return mb[:0]
}

func (mb MultiBuffer) Copy(b []byte) int {
	total := 0
	for _, bb := range mb {
		nBytes := copy(b[total:], bb.Bytes())
		total += nBytes
		if int32(nBytes) < bb.Len() {
			break
		}
	}
	return total
}

func ReadFrom(reader io.Reader) (MultiBuffer, error) {
	mb := make(MultiBuffer, 0, 16)
	for {
		b := New()
		_, err := b.ReadFullFrom(reader, Size)
		if b.IsEmpty() {
			b.Release()
		} else {
			mb = append(mb, b)
		}
		if err != nil {
			if errors.Cause(err) == io.EOF || errors.Cause(err) == io.ErrUnexpectedEOF {
				return mb, nil
			}
			return mb, err
		}
	}
}

func SplitBytes(mb MultiBuffer, b []byte) (MultiBuffer, int) {
	totalBytes := 0
	endIndex := -1
	for i := range mb {
		pBuffer := mb[i]
		nBytes, _ := pBuffer.Read(b)
		totalBytes += nBytes
		b = b[nBytes:]
		if !pBuffer.IsEmpty() {
			endIndex = i
			break
		}
		pBuffer.Release()
		mb[i] = nil
	}

	if endIndex == -1 {
		mb = mb[:0]
	} else {
		mb = mb[endIndex:]
	}

	return mb, totalBytes
}

func SplitFirstBytes(mb MultiBuffer, p []byte) (MultiBuffer, int) {
	mb, b := SplitFirst(mb)
	if b == nil {
		return mb, 0
	}
	n := copy(p, b.Bytes())
	b.Release()
	return mb, n
}

func Compact(mb MultiBuffer) MultiBuffer {
	if len(mb) == 0 {
		return mb
	}

	mb2 := make(MultiBuffer, 0, len(mb))
	last := mb[0]

	for i := 1; i < len(mb); i++ {
		curr := mb[i]
		if last.Len()+curr.Len() > Size {
			mb2 = append(mb2, last)
			last = curr
		} else {
			common.Must2(last.ReadFrom(curr))
			curr.Release()
		}
	}

	mb2 = append(mb2, last)
	return mb2
}

func SplitFirst(mb MultiBuffer) (MultiBuffer, *Buffer) {
	if len(mb) == 0 {
		return mb, nil
	}

	b := mb[0]
	mb[0] = nil
	mb = mb[1:]
	return mb, b
}

func SplitSize(mb MultiBuffer, size int32) (MultiBuffer, MultiBuffer) {
	if len(mb) == 0 {
		return mb, nil
	}

	if mb[0].Len() > size {
		b := New()
		copy(b.Extend(size), mb[0].BytesTo(size))
		mb[0].Advance(size)
		return mb, MultiBuffer{b}
	}

	totalBytes := int32(0)
	var r MultiBuffer
	endIndex := -1
	for i := range mb {
		if totalBytes+mb[i].Len() > size {
			endIndex = i
			break
		}
		totalBytes += mb[i].Len()
		r = append(r, mb[i])
		mb[i] = nil
	}
	if endIndex == -1 {
		mb = mb[:0]
	} else {
		mb = mb[endIndex:]
	}
	return mb, r
}

func WriteMultiBuffer(writer io.Writer, mb MultiBuffer) (MultiBuffer, error) {
	for {
		mb2, b := SplitFirst(mb)
		mb = mb2
		if b == nil {
			break
		}

		_, err := writer.Write(b.Bytes())
		b.Release()
		if err != nil {
			return mb, err
		}
	}

	return nil, nil
}

func (mb MultiBuffer) Len() int32 {
	if mb == nil {
		return 0
	}

	size := int32(0)
	for _, b := range mb {
		size += b.Len()
	}
	return size
}

func (mb MultiBuffer) IsEmpty() bool {
	for _, b := range mb {
		if !b.IsEmpty() {
			return false
		}
	}
	return true
}

func (mb MultiBuffer) String() string {
	v := make([]interface{}, len(mb))
	for i, b := range mb {
		v[i] = b
	}
	return serial.Concat(v...)
}

type MultiBufferContainer struct {
	MultiBuffer
}

func (c *MultiBufferContainer) Read(b []byte) (int, error) {
	if c.MultiBuffer.IsEmpty() {
		return 0, io.EOF
	}

	mb, nBytes := SplitBytes(c.MultiBuffer, b)
	c.MultiBuffer = mb
	return nBytes, nil
}

func (c *MultiBufferContainer) ReadMultiBuffer() (MultiBuffer, error) {
	mb := c.MultiBuffer
	c.MultiBuffer = nil
	return mb, nil
}

func (c *MultiBufferContainer) Write(b []byte) (int, error) {
	c.MultiBuffer = MergeBytes(c.MultiBuffer, b)
	return len(b), nil
}

func (c *MultiBufferContainer) WriteMultiBuffer(b MultiBuffer) error {
	mb, _ := MergeMulti(c.MultiBuffer, b)
	c.MultiBuffer = mb
	return nil
}

func (c *MultiBufferContainer) Close() error {
	c.MultiBuffer = ReleaseMulti(c.MultiBuffer)
	return nil
}
