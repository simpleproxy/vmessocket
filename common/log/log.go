package log

import (
	"sync"

	"github.com/vmessocket/vmessocket/common/serial"
)

type Message interface {
	String() string
}

type Handler interface {
	Handle(msg Message)
}

type GeneralMessage struct {
	Severity Severity
	Content  interface{}
}

func (m *GeneralMessage) String() string {
	return serial.Concat("[", m.Severity, "] ", m.Content)
}

func Record(msg Message) {
	logHandler.Handle(msg)
}

var logHandler syncHandler

func RegisterHandler(handler Handler) {
	if handler == nil {
		panic("Log handler is nil")
	}
	logHandler.Set(handler)
}

type syncHandler struct {
	sync.RWMutex
	Handler
}

func (h *syncHandler) Handle(msg Message) {
	h.RLock()
	defer h.RUnlock()

	if h.Handler != nil {
		h.Handler.Handle(msg)
	}
}

func (h *syncHandler) Set(handler Handler) {
	h.Lock()
	defer h.Unlock()

	h.Handler = handler
}
