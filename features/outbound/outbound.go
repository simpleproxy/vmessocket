package outbound

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/features"
	"github.com/vmessocket/vmessocket/transport"
)

type Handler interface {
	common.Runnable
	Dispatch(ctx context.Context, link *transport.Link)
}

type HandlerSelector interface {
	Select([]string) []string
}

type Manager interface {
	features.Feature
	GetHandler(tag string) Handler
	GetDefaultHandler() Handler
	AddHandler(ctx context.Context, handler Handler) error
	RemoveHandler(ctx context.Context, tag string) error
}

func ManagerType() interface{} {
	return (*Manager)(nil)
}
