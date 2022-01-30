package session

import (
	"context"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/features/routing"
)

type Context struct {
	Inbound  *session.Inbound
	Outbound *session.Outbound
	Content  *session.Content
}

func (ctx *Context) GetInboundTag() string {
	if ctx.Inbound == nil {
		return ""
	}
	return ctx.Inbound.Tag
}

func (ctx *Context) GetSourceIPs() []net.IP {
	if ctx.Inbound == nil || !ctx.Inbound.Source.IsValid() {
		return nil
	}
	dest := ctx.Inbound.Source
	if dest.Address.Family().IsDomain() {
		return nil
	}

	return []net.IP{dest.Address.IP()}
}

func (ctx *Context) GetSourcePort() net.Port {
	if ctx.Inbound == nil || !ctx.Inbound.Source.IsValid() {
		return 0
	}
	return ctx.Inbound.Source.Port
}

func (ctx *Context) GetTargetIPs() []net.IP {
	if ctx.Outbound == nil || !ctx.Outbound.Target.IsValid() {
		return nil
	}

	if ctx.Outbound.Target.Address.Family().IsIP() {
		return []net.IP{ctx.Outbound.Target.Address.IP()}
	}

	return nil
}

func (ctx *Context) GetTargetPort() net.Port {
	if ctx.Outbound == nil || !ctx.Outbound.Target.IsValid() {
		return 0
	}
	return ctx.Outbound.Target.Port
}

func (ctx *Context) GetTargetDomain() string {
	if ctx.Outbound == nil || !ctx.Outbound.Target.IsValid() {
		return ""
	}
	dest := ctx.Outbound.Target
	if !dest.Address.Family().IsDomain() {
		return ""
	}
	return dest.Address.Domain()
}

func (ctx *Context) GetNetwork() net.Network {
	if ctx.Outbound == nil {
		return net.Network_Unknown
	}
	return ctx.Outbound.Target.Network
}

func (ctx *Context) GetProtocol() string {
	if ctx.Content == nil {
		return ""
	}
	return ctx.Content.Protocol
}

func (ctx *Context) GetUser() string {
	if ctx.Inbound == nil || ctx.Inbound.User == nil {
		return ""
	}
	return ctx.Inbound.User.Email
}

func (ctx *Context) GetAttributes() map[string]string {
	if ctx.Content == nil {
		return nil
	}
	return ctx.Content.Attributes
}

func (ctx *Context) GetSkipDNSResolve() bool {
	if ctx.Content == nil {
		return false
	}
	return ctx.Content.SkipDNSResolve
}

func AsRoutingContext(ctx context.Context) routing.Context {
	return &Context{
		Inbound:  session.InboundFromContext(ctx),
		Outbound: session.OutboundFromContext(ctx),
		Content:  session.ContentFromContext(ctx),
	}
}
