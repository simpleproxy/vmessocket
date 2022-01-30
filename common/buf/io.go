package buf

import (
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

type Reader interface {
	ReadMultiBuffer() (MultiBuffer, error)
}

var ErrReadTimeout = newError("IO timeout")

type TimeoutReader interface {
	ReadMultiBufferTimeout(time.Duration) (MultiBuffer, error)
}

type Writer interface {
	WriteMultiBuffer(MultiBuffer) error
}

func WriteAllBytes(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		n, err := writer.Write(payload)
		if err != nil {
			return err
		}
		payload = payload[n:]
	}
	return nil
}

func isPacketReader(reader io.Reader) bool {
	_, ok := reader.(net.PacketConn)
	return ok
}

func NewReader(reader io.Reader) Reader {
	if mr, ok := reader.(Reader); ok {
		return mr
	}

	if isPacketReader(reader) {
		return &PacketReader{
			Reader: reader,
		}
	}

	_, isFile := reader.(*os.File)
	if !isFile && useReadv {
		if sc, ok := reader.(syscall.Conn); ok {
			rawConn, err := sc.SyscallConn()
			if err != nil {
				newError("failed to get sysconn").Base(err).WriteToLog()
			} else {
				return NewReadVReader(reader, rawConn)
			}
		}
	}

	return &SingleReader{
		Reader: reader,
	}
}

func NewPacketReader(reader io.Reader) Reader {
	if mr, ok := reader.(Reader); ok {
		return mr
	}

	return &PacketReader{
		Reader: reader,
	}
}

func isPacketWriter(writer io.Writer) bool {
	if _, ok := writer.(net.PacketConn); ok {
		return true
	}

	if _, ok := writer.(syscall.Conn); !ok {
		return true
	}
	return false
}

func NewWriter(writer io.Writer) Writer {
	if mw, ok := writer.(Writer); ok {
		return mw
	}

	if isPacketWriter(writer) {
		return &SequentialWriter{
			Writer: writer,
		}
	}

	return &BufferToBytesWriter{
		Writer: writer,
	}
}
