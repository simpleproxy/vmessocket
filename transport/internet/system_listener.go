package internet

import (
	"context"
	"runtime"
	"syscall"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
)

var effectiveListener = DefaultListener{}

type controller func(network, address string, fd uintptr) error

type DefaultListener struct {
	controllers []controller
}

func getControlFunc(ctx context.Context, sockopt *SocketConfig, controllers []controller) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			for _, controller := range controllers {
				if err := controller(network, address, fd); err != nil {
					newError("failed to apply external controller").Base(err).WriteToLog(session.ExportIDToError(ctx))
				}
			}
		})
	}
}

func RegisterListenerController(controller func(network, address string, fd uintptr) error) error {
	if controller == nil {
		return newError("nil listener controller")
	}
	effectiveListener.controllers = append(effectiveListener.controllers, controller)
	return nil
}

func (dl *DefaultListener) Listen(ctx context.Context, addr net.Addr, sockopt *SocketConfig) (net.Listener, error) {
	var lc net.ListenConfig
	var l net.Listener
	var err error
	var network, address string
	switch addr := addr.(type) {
	case *net.TCPAddr:
		network = addr.Network()
		address = addr.String()
		lc.Control = getControlFunc(ctx, sockopt, dl.controllers)
	case *net.UnixAddr:
		lc.Control = nil
		network = addr.Network()
		address = addr.Name
		if (runtime.GOOS == "linux" || runtime.GOOS == "android") && address[0] == '@' {
			if len(address) > 1 && address[1] == '@' {
				fullAddr := make([]byte, len(syscall.RawSockaddrUnix{}.Path))
				copy(fullAddr, address[1:])
				address = string(fullAddr)
			}
		}
	}
	l, err = lc.Listen(ctx, network, address)
	return l, err
}

func (dl *DefaultListener) ListenPacket(ctx context.Context, addr net.Addr, sockopt *SocketConfig) (net.PacketConn, error) {
	var lc net.ListenConfig
	lc.Control = getControlFunc(ctx, sockopt, dl.controllers)
	return lc.ListenPacket(ctx, addr.Network(), addr.String())
}
