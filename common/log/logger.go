package log

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/common/signal/semaphore"
)

type consoleLogWriter struct {
	logger *log.Logger
}

type fileLogWriter struct {
	file   *os.File
	logger *log.Logger
}

type generalLogger struct {
	creator WriterCreator
	buffer  chan Message
	access  *semaphore.Instance
	done    *done.Instance
}

type Writer interface {
	Write(string) error
	io.Closer
}

type WriterCreator func() Writer

func CreateFileLogWriter(path string) (WriterCreator, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	file.Close()
	return func() Writer {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return nil
		}
		return &fileLogWriter{
			file:   file,
			logger: log.New(file, "", log.Ldate|log.Ltime),
		}
	}, nil
}

func CreateStderrLogWriter() WriterCreator {
	return func() Writer {
		return &consoleLogWriter{
			logger: log.New(os.Stderr, "", log.Ldate|log.Ltime),
		}
	}
}

func CreateStdoutLogWriter() WriterCreator {
	return func() Writer {
		return &consoleLogWriter{
			logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		}
	}
}

func NewLogger(logWriterCreator WriterCreator) Handler {
	return &generalLogger{
		creator: logWriterCreator,
		buffer:  make(chan Message, 16),
		access:  semaphore.New(1),
		done:    done.New(),
	}
}

func (w *consoleLogWriter) Close() error {
	return nil
}

func (w *fileLogWriter) Close() error {
	return w.file.Close()
}

func (l *generalLogger) Close() error {
	return l.done.Close()
}

func (l *generalLogger) Handle(msg Message) {
	select {
	case l.buffer <- msg:
	default:
	}
	select {
	case <-l.access.Wait():
		go l.run()
	default:
	}
}

func (l *generalLogger) run() {
	defer l.access.Signal()
	dataWritten := false
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	logger := l.creator()
	if logger == nil {
		return
	}
	defer logger.Close()
	for {
		select {
		case <-l.done.Wait():
			return
		case msg := <-l.buffer:
			logger.Write(msg.String() + platform.LineSeparator())
			dataWritten = true
		case <-ticker.C:
			if !dataWritten {
				return
			}
			dataWritten = false
		}
	}
}

func (w *consoleLogWriter) Write(s string) error {
	w.logger.Print(s)
	return nil
}

func (w *fileLogWriter) Write(s string) error {
	w.logger.Print(s)
	return nil
}

func init() {
	RegisterHandler(NewLogger(CreateStdoutLogWriter()))
}
