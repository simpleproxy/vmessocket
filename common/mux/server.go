package mux

import (
	"context"

	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/routing"
	"github.com/vmessocket/vmessocket/transport"
)

type Server struct {
	dispatcher routing.Dispatcher
}

type ServerWorker struct {
	dispatcher     routing.Dispatcher
	link           *transport.Link
	sessionManager *SessionManager
}

func NewServer(ctx context.Context) *Server {
	s := &Server{}
	core.RequireFeatures(ctx, func(d routing.Dispatcher) {
		s.dispatcher = d
	})
	return s
}

func (w *ServerWorker) ActiveConnections() uint32 {
	return uint32(w.sessionManager.Size())
}

func (s *Server) Close() error {
	return nil
}

func (w *ServerWorker) Closed() bool {
	return w.sessionManager.Closed()
}

func (w *ServerWorker) handleFrame(ctx context.Context, reader *buf.BufferedReader) error {
	var meta FrameMetadata
	err := meta.Unmarshal(reader)
	if err != nil {
		return newError("failed to read metadata").Base(err)
	}
	switch meta.SessionStatus {
	default:
		status := meta.SessionStatus
		return newError("unknown status: ", status).AtError()
	}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Type() interface{} {
	return s.dispatcher.Type()
}
