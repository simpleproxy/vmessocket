package pipe

import (
	"github.com/vmessocket/vmessocket/common/buf"
)

type Writer struct {
	pipe *pipe
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	return w.pipe.WriteMultiBuffer(mb)
}

func (w *Writer) Close() error {
	return w.pipe.Close()
}

func (w *Writer) Interrupt() {
	w.pipe.Interrupt()
}
