package mux

import (
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/signal/done"
)

var muxCoolAddress = net.DomainAddress("v1.mux.cool")

type ClientWorker struct {
	sessionManager *SessionManager
	done           *done.Instance
}

type WorkerPicker interface {
	PickAvailable() (*ClientWorker, error)
}

func (m *ClientWorker) Closed() bool {
	return m.done.Done()
}
