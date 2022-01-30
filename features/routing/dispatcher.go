package routing

import (
	"context"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features"
	"github.com/vmessocket/vmessocket/transport"
)

type Dispatcher interface {
	features.Feature
	Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error)
}

func DispatcherType() interface{} {
	return (*Dispatcher)(nil)
}
