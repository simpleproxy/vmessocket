package commander

import (
	"context"
	"sync"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/transport"
)

type OutboundListener struct {
	buffer chan net.Conn
	done   *done.Instance
}

func (l *OutboundListener) add(conn net.Conn) {
	select {
	case l.buffer <- conn:
	case <-l.done.Wait():
		conn.Close()
	default:
		conn.Close()
	}
}

func (l *OutboundListener) Accept() (net.Conn, error) {
	select {
	case <-l.done.Wait():
		return nil, newError("listen closed")
	case c := <-l.buffer:
		return c, nil
	}
}

func (l *OutboundListener) Close() error {
	common.Must(l.done.Close())
L:
	for {
		select {
		case c := <-l.buffer:
			c.Close()
		default:
			break L
		}
	}
	return nil
}

func (l *OutboundListener) Addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: 0,
	}
}

type Outbound struct {
	tag      string
	listener *OutboundListener
	access   sync.RWMutex
	closed   bool
}

func (co *Outbound) Dispatch(ctx context.Context, link *transport.Link) {
	co.access.RLock()

	if co.closed {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		co.access.RUnlock()
		return
	}

	closeSignal := done.New()
	c := net.NewConnection(net.ConnectionInputMulti(link.Writer), net.ConnectionOutputMulti(link.Reader), net.ConnectionOnClose(closeSignal))
	co.listener.add(c)
	co.access.RUnlock()
	<-closeSignal.Wait()
}

func (co *Outbound) Tag() string {
	return co.tag
}

func (co *Outbound) Start() error {
	co.access.Lock()
	co.closed = false
	co.access.Unlock()
	return nil
}

func (co *Outbound) Close() error {
	co.access.Lock()
	defer co.access.Unlock()

	co.closed = true
	return co.listener.Close()
}
