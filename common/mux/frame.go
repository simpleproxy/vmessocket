package mux

import (
	"encoding/binary"

	"github.com/vmessocket/vmessocket/common/bitmask"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/protocol"
)

const (
	OptionData             bitmask.Byte  = 0x01
	OptionError            bitmask.Byte  = 0x02
	SessionStatusNew       SessionStatus = 0x01
	SessionStatusKeep      SessionStatus = 0x02
	SessionStatusEnd       SessionStatus = 0x03
	SessionStatusKeepAlive SessionStatus = 0x04
	TargetNetworkTCP       TargetNetwork = 0x01
	TargetNetworkUDP       TargetNetwork = 0x02
)

var addrParser = protocol.NewAddressParser(
	protocol.AddressFamilyByte(byte(protocol.AddressTypeIPv4), net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(byte(protocol.AddressTypeDomain), net.AddressFamilyDomain),
	protocol.AddressFamilyByte(byte(protocol.AddressTypeIPv6), net.AddressFamilyIPv6),
	protocol.PortThenAddress(),
)

type FrameMetadata struct {
	Target        net.Destination
	SessionID     uint16
	Option        bitmask.Byte
	SessionStatus SessionStatus
}

type SessionStatus byte

type TargetNetwork byte

func (f *FrameMetadata) UnmarshalFromBuffer(b *buf.Buffer) error {
	if b.Len() < 4 {
		return newError("insufficient buffer: ", b.Len())
	}
	f.SessionID = binary.BigEndian.Uint16(b.BytesTo(2))
	f.SessionStatus = SessionStatus(b.Byte(2))
	f.Option = bitmask.Byte(b.Byte(3))
	f.Target.Network = net.Network_Unknown
	if f.SessionStatus == SessionStatusNew {
		if b.Len() < 8 {
			return newError("insufficient buffer: ", b.Len())
		}
		network := TargetNetwork(b.Byte(4))
		b.Advance(5)
		addr, port, err := addrParser.ReadAddressPort(nil, b)
		if err != nil {
			return newError("failed to parse address and port").Base(err)
		}
		switch network {
		case TargetNetworkTCP:
			f.Target = net.TCPDestination(addr, port)
		case TargetNetworkUDP:
			f.Target = net.UDPDestination(addr, port)
		default:
			return newError("unknown network type: ", network)
		}
	}
	return nil
}
