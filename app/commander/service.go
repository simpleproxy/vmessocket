package commander

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/vmessocket/vmessocket/common"
)

type reflectionService struct{}

type Service interface {
	Register(*grpc.Server)
}

func (r reflectionService) Register(s *grpc.Server) {
	reflection.Register(s)
}

func init() {
	common.Must(common.RegisterConfig((*ReflectionConfig)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		return reflectionService{}, nil
	}))
}
