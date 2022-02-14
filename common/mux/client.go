package mux

import (
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/transport"
)

var (
	muxCoolAddress = net.DomainAddress("v1.mux.cool")
	muxCoolPort    = net.Port(9527)
)

type ClientWorker struct {
	sessionManager *SessionManager
	link           transport.Link
	done           *done.Instance
}

type WorkerPicker interface {
	PickAvailable() (*ClientWorker, error)
}

func (m *ClientWorker) Closed() bool {
	return m.done.Done()
}

func (m *ClientWorker) TotalConnections() uint32 {
	return uint32(m.sessionManager.Count())
}
