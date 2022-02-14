package mux

import (
	"github.com/vmessocket/vmessocket/common/net"
)

var muxCoolAddress = net.DomainAddress("v1.mux.cool")

type ClientWorker struct {
}

type WorkerPicker interface {
	PickAvailable() (*ClientWorker, error)
}
