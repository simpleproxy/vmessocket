package udp

import (
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/net"
)

type Packet struct {
	Payload *buf.Buffer
	Source  net.Destination
	Target  net.Destination
}
