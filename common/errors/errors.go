package errors

import (
	"os"
	"reflect"
	"strings"

	"github.com/vmessocket/vmessocket/common/log"
	"github.com/vmessocket/vmessocket/common/serial"
)

type Error struct {
	pathObj  interface{}
	prefix   []interface{}
	message  []interface{}
	inner    error
	severity log.Severity
}

type ExportOption func(*ExportOptionHolder)

type ExportOptionHolder struct {
	SessionID uint32
}

type hasInnerError interface {
	Inner() error
}

type hasSeverity interface {
	Severity() log.Severity
}

func Cause(err error) error {
	if err == nil {
		return nil
	}
L:
	for {
		switch inner := err.(type) {
		case hasInnerError:
			if inner.Inner() == nil {
				break L
			}
			err = inner.Inner()
		case *os.PathError:
			if inner.Err == nil {
				break L
			}
			err = inner.Err
		case *os.SyscallError:
			if inner.Err == nil {
				break L
			}
			err = inner.Err
		default:
			break L
		}
	}
	return err
}

func GetSeverity(err error) log.Severity {
	if s, ok := err.(hasSeverity); ok {
		return s.Severity()
	}
	return log.Severity_Info
}

func New(msg ...interface{}) *Error {
	return &Error{
		message:  msg,
		severity: log.Severity_Info,
	}
}

func (err *Error) AtDebug() *Error {
	return err.atSeverity(log.Severity_Debug)
}

func (err *Error) AtError() *Error {
	return err.atSeverity(log.Severity_Error)
}

func (err *Error) AtInfo() *Error {
	return err.atSeverity(log.Severity_Info)
}

func (err *Error) atSeverity(s log.Severity) *Error {
	err.severity = s
	return err
}

func (err *Error) AtWarning() *Error {
	return err.atSeverity(log.Severity_Warning)
}

func (err *Error) Base(e error) *Error {
	err.inner = e
	return err
}

func (err *Error) Error() string {
	builder := strings.Builder{}
	for _, prefix := range err.prefix {
		builder.WriteByte('[')
		builder.WriteString(serial.ToString(prefix))
		builder.WriteString("] ")
	}
	path := err.pkgPath()
	if len(path) > 0 {
		builder.WriteString(path)
		builder.WriteString(": ")
	}
	msg := serial.Concat(err.message...)
	builder.WriteString(msg)
	if err.inner != nil {
		builder.WriteString(" > ")
		builder.WriteString(err.inner.Error())
	}
	return builder.String()
}

func (err *Error) Inner() error {
	if err.inner == nil {
		return nil
	}
	return err.inner
}

func (err *Error) pkgPath() string {
	if err.pathObj == nil {
		return ""
	}
	path := reflect.TypeOf(err.pathObj).PkgPath()
	path = strings.TrimPrefix(path, "github.com/vmessocket/vmessocket/")
	path = strings.TrimPrefix(path, "github.com/vmessocket/vmessocket")
	return path
}

func (err *Error) Severity() log.Severity {
	if err.inner == nil {
		return err.severity
	}
	if s, ok := err.inner.(hasSeverity); ok {
		as := s.Severity()
		if as < err.severity {
			return as
		}
	}
	return err.severity
}

func (err *Error) String() string {
	return err.Error()
}

func (err *Error) WithPathObj(obj interface{}) *Error {
	err.pathObj = obj
	return err
}

func (err *Error) WriteToLog(opts ...ExportOption) {
	var holder ExportOptionHolder
	for _, opt := range opts {
		opt(&holder)
	}
	if holder.SessionID > 0 {
		err.prefix = append(err.prefix, holder.SessionID)
	}
	log.Record(&log.GeneralMessage{
		Severity: GetSeverity(err),
		Content:  err,
	})
}
