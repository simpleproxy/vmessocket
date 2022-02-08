package routing

import "github.com/vmessocket/vmessocket/common/net"

type Context interface {
	GetInboundTag() string
	GetSourceIPs() []net.IP
	GetSourcePort() net.Port
	GetTargetIPs() []net.IP
	GetTargetPort() net.Port
	GetTargetDomain() string
	GetNetwork() net.Network
	GetProtocol() string
	GetUser() string
	GetAttributes() map[string]string
	GetSkipDNSResolve() bool
}
