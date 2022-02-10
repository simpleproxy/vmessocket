package inbound

import (
	"context"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/routing"
)

type Server struct {
	dispatcher routing.Dispatcher
}

func NewServer(ctx context.Context) *Server {
	s := &Server{}
	core.RequireFeatures(ctx, func(d routing.Dispatcher) {
		s.dispatcher = d
	})
	return s
}
