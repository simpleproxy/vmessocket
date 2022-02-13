package mux

import (
	"context"
	"io"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
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

func handle(ctx context.Context, s *Session, output buf.Writer) {
	writer := NewResponseWriter(s.ID, output, s.transferType)
	if err := buf.Copy(s.input, writer); err != nil {
		newError("session ", s.ID, " ends.").Base(err).WriteToLog(session.ExportIDToError(ctx))
		writer.hasError = true
	}
	writer.Close()
	s.Close()
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
	go worker.run(ctx)
	return worker, nil
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

func (w *ServerWorker) run(ctx context.Context) {
	input := w.link.Reader
	reader := &buf.BufferedReader{Reader: input}
	defer w.sessionManager.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := w.handleFrame(ctx, reader)
			if err != nil {
				if errors.Cause(err) != io.EOF {
					newError("unexpected EOF").Base(err).WriteToLog(session.ExportIDToError(ctx))
					common.Interrupt(input)
				}
				return
			}
		}
	}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Type() interface{} {
	return s.dispatcher.Type()
}
