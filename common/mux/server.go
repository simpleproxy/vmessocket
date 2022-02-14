package mux

import (
	"context"

	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/routing"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/pipe"
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

func NewServerWorker(ctx context.Context, d routing.Dispatcher, link *transport.Link) (*ServerWorker, error) {
	worker := &ServerWorker{
		dispatcher:     d,
		link:           link,
		sessionManager: NewSessionManager(),
	}
	return worker, nil
}

func (s *Server) Close() error {
	return nil
}

func (w *ServerWorker) Closed() bool {
	return w.sessionManager.Closed()
}

func (s *Server) Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error) {
	if dest.Address != muxCoolAddress {
		return s.dispatcher.Dispatch(ctx, dest)
	}
	opts := pipe.OptionsFromContext(ctx)
	uplinkReader, uplinkWriter := pipe.New(opts...)
	downlinkReader, downlinkWriter := pipe.New(opts...)
	_, err := NewServerWorker(ctx, s.dispatcher, &transport.Link{
		Reader: uplinkReader,
		Writer: downlinkWriter,
	})
	if err != nil {
		return nil, err
	}
	return &transport.Link{Reader: downlinkReader, Writer: uplinkWriter}, nil
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
