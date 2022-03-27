package internet

import (
	"context"
	"time"

	"github.com/vmessocket/vmessocket/common/net"
)

var effectiveSystemDialer SystemDialer = &DefaultSystemDialer{}

type DefaultSystemDialer struct {
	controllers []controller
}

type packetConnWrapper struct {
	conn net.PacketConn
	dest net.Addr
}

type SimpleSystemDialer struct {
	adapter SystemDialerAdapter
}

type SystemDialer interface {
	Dial(ctx context.Context, source net.Address, destination net.Destination, sockopt *SocketConfig) (net.Conn, error)
}

type SystemDialerAdapter interface {
	Dial(network string, address string) (net.Conn, error)
}

func hasBindAddr(sockopt *SocketConfig) bool {
	return sockopt != nil && len(sockopt.BindAddress) > 0 && sockopt.BindPort > 0
}

func RegisterDialerController(ctl func(network, address string, fd uintptr) error) error {
	if ctl == nil {
		return newError("nil listener controller")
	}
	dialer, ok := effectiveSystemDialer.(*DefaultSystemDialer)
	if !ok {
		return newError("RegisterListenerController not supported in custom dialer")
	}
	dialer.controllers = append(dialer.controllers, ctl)
	return nil
}

func resolveSrcAddr(network net.Network, src net.Address) net.Addr {
	if src == nil || src == net.AnyIP {
		return nil
	}
	if network == net.Network_TCP {
		return &net.TCPAddr{
			IP:   src.IP(),
			Port: 0,
		}
	}
	return &net.UDPAddr{
		IP:   src.IP(),
		Port: 0,
	}
}

func UseAlternativeSystemDialer(dialer SystemDialer) {
	if dialer == nil {
		dialer = &DefaultSystemDialer{}
	}
	effectiveSystemDialer = dialer
}

func WithAdapter(dialer SystemDialerAdapter) SystemDialer {
	return &SimpleSystemDialer{
		adapter: dialer,
	}
}

func (c *packetConnWrapper) Close() error {
	return c.conn.Close()
}

func (d *DefaultSystemDialer) Dial(ctx context.Context, src net.Address, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	if dest.Network == net.Network_UDP && !hasBindAddr(sockopt) {
		srcAddr := resolveSrcAddr(net.Network_UDP, src)
		if srcAddr == nil {
			srcAddr = &net.UDPAddr{
				IP:   []byte{0, 0, 0, 0},
				Port: 0,
			}
		}
		packetConn, err := ListenSystemPacket(ctx, srcAddr, sockopt)
		if err != nil {
			return nil, err
		}
		destAddr, err := net.ResolveUDPAddr("udp", dest.NetAddr())
		if err != nil {
			return nil, err
		}
		return &packetConnWrapper{
			conn: packetConn,
			dest: destAddr,
		}, nil
	}
	dialer := &net.Dialer{
		Timeout:   time.Second * 16,
		LocalAddr: resolveSrcAddr(dest.Network, src),
	}
	return dialer.DialContext(ctx, dest.Network.SystemString(), dest.NetAddr())
}

func (v *SimpleSystemDialer) Dial(ctx context.Context, src net.Address, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	return v.adapter.Dial(dest.Network.SystemString(), dest.NetAddr())
}

func (c *packetConnWrapper) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *packetConnWrapper) Read(p []byte) (int, error) {
	n, _, err := c.conn.ReadFrom(p)
	return n, err
}

func (c *packetConnWrapper) RemoteAddr() net.Addr {
	return c.dest
}

func (c *packetConnWrapper) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *packetConnWrapper) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *packetConnWrapper) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *packetConnWrapper) Write(p []byte) (int, error) {
	return c.conn.WriteTo(p, c.dest)
}
