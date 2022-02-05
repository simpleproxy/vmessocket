package command

import (
	"context"

	"google.golang.org/grpc"

	"github.com/vmessocket/vmessocket/app/log"
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/core"
)

type LoggerServer struct {
	V *core.Instance
}

type service struct {
	v *core.Instance
}

func (s *LoggerServer) mustEmbedUnimplementedLoggerServiceServer() {}

func (s *service) Register(server *grpc.Server) {
	RegisterLoggerServiceServer(server, &LoggerServer{
		V: s.v,
	})
}

func (s *LoggerServer) RestartLogger(ctx context.Context, request *RestartLoggerRequest) (*RestartLoggerResponse, error) {
	logger := s.V.GetFeature((*log.Instance)(nil))
	if logger == nil {
		return nil, newError("unable to get logger instance")
	}
	if err := logger.Close(); err != nil {
		return nil, newError("failed to close logger").Base(err)
	}
	if err := logger.Start(); err != nil {
		return nil, newError("failed to start logger").Base(err)
	}
	return &RestartLoggerResponse{}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		s := core.MustFromContext(ctx)
		return &service{v: s}, nil
	}))
}
