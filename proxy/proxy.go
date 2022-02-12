package proxy

import (
	"context"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/protocol"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/internet"
)

type GetInbound interface {
	GetInbound() Inbound
}

type GetOutbound interface {
	GetOutbound() Outbound
}

type Inbound interface {
	Network() []net.Network
	Process(context.Context, net.Network, internet.Connection) error
}

type Outbound interface {
	Process(context.Context, *transport.Link, internet.Dialer) error
}

type UserManager interface {
	AddUser(context.Context, *protocol.MemoryUser) error
	RemoveUser(context.Context, string) error
}
