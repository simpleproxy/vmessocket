package log

import (
	"sync"

	"github.com/vmessocket/vmessocket/common/serial"
)

var logHandler syncHandler

type GeneralMessage struct {
	Severity Severity
	Content  interface{}
}

type Handler interface {
	Handle(msg Message)
}

type Message interface {
	String() string
}

type syncHandler struct {
	sync.RWMutex
	Handler
}

func Record(msg Message) {
	logHandler.Handle(msg)
}

func RegisterHandler(handler Handler) {
	if handler == nil {
		panic("Log handler is nil")
	}
	logHandler.Set(handler)
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

func (m *GeneralMessage) String() string {
	return serial.Concat("[", m.Severity, "] ", m.Content)
}
