package routing

import "github.com/vmessocket/vmessocket/common/net"

type Context interface {
	GetAttributes() map[string]string
	GetInboundTag() string
	GetNetwork() net.Network
	GetProtocol() string
	GetSkipDNSResolve() bool
	GetSourceIPs() []net.IP
	GetSourcePort() net.Port
	GetTargetDomain() string
	GetTargetIPs() []net.IP
	GetTargetPort() net.Port
	GetUser() string
}
