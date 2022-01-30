package inbound

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features"
)

type Handler interface {
	common.Runnable
	Tag() string
	GetRandomInboundProxy() (interface{}, net.Port, int)
}

type Manager interface {
	features.Feature
	GetHandler(ctx context.Context, tag string) (Handler, error)
	AddHandler(ctx context.Context, handler Handler) error
	RemoveHandler(ctx context.Context, tag string) error
}

func ManagerType() interface{} {
	return (*Manager)(nil)
}
