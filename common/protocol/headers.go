package protocol

import (
	"runtime"

	"github.com/vmessocket/vmessocket/common/bitmask"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/uuid"
)

const (
	RequestCommandTCP = RequestCommand(0x01)
	RequestCommandUDP = RequestCommand(0x02)
	RequestOptionChunkStream         bitmask.Byte = 0x01
	RequestOptionConnectionReuse     bitmask.Byte = 0x02
	RequestOptionChunkMasking        bitmask.Byte = 0x04
	RequestOptionGlobalPadding       bitmask.Byte = 0x08
	RequestOptionAuthenticatedLength bitmask.Byte = 0x10
	ResponseOptionConnectionReuse bitmask.Byte = 0x01
)

type CommandSwitchAccount struct {
	Host     net.Address
	Port     net.Port
	ID       uuid.UUID
	Level    uint32
	AlterIds uint16
	ValidMin byte
}

type RequestCommand byte

type RequestHeader struct {
	Version  byte
	Command  RequestCommand
	Option   bitmask.Byte
	Security SecurityType
	Port     net.Port
	Address  net.Address
	User     *MemoryUser
}

type ResponseCommand interface{}

type ResponseHeader struct {
	Option  bitmask.Byte
	Command ResponseCommand
}

func isDomainTooLong(domain string) bool {
	return len(domain) > 256
}

func (h *RequestHeader) Destination() net.Destination {
	if h.Command == RequestCommandUDP {
		return net.UDPDestination(h.Address, h.Port)
	}
	return net.TCPDestination(h.Address, h.Port)
}

func (sc *SecurityConfig) GetSecurityType() SecurityType {
	if sc == nil || sc.Type == SecurityType_AUTO {
		if runtime.GOARCH == "amd64" || runtime.GOARCH == "s390x" || runtime.GOARCH == "arm64" {
			return SecurityType_AES128_GCM
		}
		return SecurityType_CHACHA20_POLY1305
	}
	return sc.Type
}

func (c RequestCommand) TransferType() TransferType {
	switch c {
	case RequestCommandTCP:
		return TransferTypeStream
	case RequestCommandUDP:
		return TransferTypePacket
	default:
		return TransferTypeStream
	}
}
