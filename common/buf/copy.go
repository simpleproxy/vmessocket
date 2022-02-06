package buf

import (
	"io"
	"time"

	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/signal"
)

var ErrNotTimeoutReader = newError("not a TimeoutReader")

type copyHandler struct {
	onData []dataHandler
}

type CopyOption func(*copyHandler)

type dataHandler func(MultiBuffer)

type readError struct {
	error
}

type SizeCounter struct {
	Size int64
}

type writeError struct {
	error
}

func Copy(reader Reader, writer Writer, options ...CopyOption) error {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	err := copyInternal(reader, writer, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return err
	}
	return nil
}

func copyInternal(reader Reader, writer Writer, handler *copyHandler) error {
	for {
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			for _, handler := range handler.onData {
				handler(buffer)
			}
			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return writeError{werr}
			}
		}
		if err != nil {
			return readError{err}
		}
	}
}

func CopyOnceTimeout(reader Reader, writer Writer, timeout time.Duration) error {
	timeoutReader, ok := reader.(TimeoutReader)
	if !ok {
		return ErrNotTimeoutReader
	}
	mb, err := timeoutReader.ReadMultiBufferTimeout(timeout)
	if err != nil {
		return err
	}
	return writer.WriteMultiBuffer(mb)
}

func CountSize(sc *SizeCounter) CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(b MultiBuffer) {
			sc.Size += int64(b.Len())
		})
	}
}

func IsReadError(err error) bool {
	_, ok := err.(readError)
	return ok
}

func IsWriteError(err error) bool {
	_, ok := err.(writeError)
	return ok
}

func UpdateActivity(timer signal.ActivityUpdater) CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(MultiBuffer) {
			timer.Update()
		})
	}
}

func (e readError) Error() string {
	return e.error.Error()
}

func (e writeError) Error() string {
	return e.error.Error()
}

func (e readError) Inner() error {
	return e.error
}

func (e writeError) Inner() error {
	return e.error
}
