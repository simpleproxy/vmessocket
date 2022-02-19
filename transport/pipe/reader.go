package pipe

import (
	"time"

	"github.com/vmessocket/vmessocket/common/buf"
)

type Reader struct {
	pipe *pipe
}

func (r *Reader) Interrupt() {
	r.pipe.Interrupt()
}

func (r *Reader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return r.pipe.ReadMultiBuffer()
}

func (r *Reader) ReadMultiBufferTimeout(d time.Duration) (buf.MultiBuffer, error) {
	return r.pipe.ReadMultiBufferTimeout(d)
}
