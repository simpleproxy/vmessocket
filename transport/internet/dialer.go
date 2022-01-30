package internet

import (
	"context"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
)

type Dialer interface {
	Dial(ctx context.Context, destination net.Destination) (Connection, error)
	Address() net.Address
}

type dialFunc func(ctx context.Context, dest net.Destination, streamSettings *MemoryStreamConfig) (Connection, error)

var transportDialerCache = make(map[string]dialFunc)

func RegisterTransportDialer(protocol string, dialer dialFunc) error {
	if _, found := transportDialerCache[protocol]; found {
		return newError(protocol, " dialer already registered").AtError()
	}
	transportDialerCache[protocol] = dialer
	return nil
}

func Dial(ctx context.Context, dest net.Destination, streamSettings *MemoryStreamConfig) (Connection, error) {
	if dest.Network == net.Network_TCP {
		if streamSettings == nil {
			s, err := ToMemoryStreamConfig(nil)
			if err != nil {
				return nil, newError("failed to create default stream settings").Base(err)
			}
			streamSettings = s
		}

		protocol := streamSettings.ProtocolName
		dialer := transportDialerCache[protocol]
		if dialer == nil {
			return nil, newError(protocol, " dialer not registered").AtError()
		}
		return dialer(ctx, dest, streamSettings)
	}

	if dest.Network == net.Network_UDP {
		udpDialer := transportDialerCache["udp"]
		if udpDialer == nil {
			return nil, newError("UDP dialer not registered").AtError()
		}
		return udpDialer(ctx, dest, streamSettings)
	}

	return nil, newError("unknown network ", dest.Network)
}

func DialSystem(ctx context.Context, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	var src net.Address
	if outbound := session.OutboundFromContext(ctx); outbound != nil {
		src = outbound.Gateway
	}

	return effectiveSystemDialer.Dial(ctx, src, dest, sockopt)
}
