package net

import (
	"net"
	"strings"
)

type Destination struct {
	Address Address
	Port    Port
	Network Network
}

func DestinationFromAddr(addr net.Addr) Destination {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		return TCPDestination(IPAddress(addr.IP), Port(addr.Port))
	case *net.UDPAddr:
		return UDPDestination(IPAddress(addr.IP), Port(addr.Port))
	case *net.UnixAddr:
		return UnixDestination(DomainAddress(addr.Name))
	default:
		panic("Net: Unknown address type.")
	}
}

func ParseDestination(dest string) (Destination, error) {
	d := Destination{
		Address: AnyIP,
		Port:    Port(0),
	}

	switch {
	case strings.HasPrefix(dest, "tcp:"):
		d.Network = Network_TCP
		dest = dest[4:]
	case strings.HasPrefix(dest, "udp:"):
		d.Network = Network_UDP
		dest = dest[4:]
	case strings.HasPrefix(dest, "unix:"):
		d = UnixDestination(DomainAddress(dest[5:]))
		return d, nil
	}

	hstr, pstr, err := SplitHostPort(dest)
	if err != nil {
		return d, err
	}
	if len(hstr) > 0 {
		d.Address = ParseAddress(hstr)
	}
	if len(pstr) > 0 {
		port, err := PortFromString(pstr)
		if err != nil {
			return d, err
		}
		d.Port = port
	}
	return d, nil
}

func TCPDestination(address Address, port Port) Destination {
	return Destination{
		Network: Network_TCP,
		Address: address,
		Port:    port,
	}
}

func UDPDestination(address Address, port Port) Destination {
	return Destination{
		Network: Network_UDP,
		Address: address,
		Port:    port,
	}
}

func UnixDestination(address Address) Destination {
	return Destination{
		Network: Network_UNIX,
		Address: address,
	}
}

func (d Destination) NetAddr() string {
	addr := ""
	if d.Network == Network_TCP || d.Network == Network_UDP {
		addr = d.Address.String() + ":" + d.Port.String()
	} else if d.Network == Network_UNIX {
		addr = d.Address.String()
	}
	return addr
}

func (d Destination) String() string {
	prefix := "unknown:"
	switch d.Network {
	case Network_TCP:
		prefix = "tcp:"
	case Network_UDP:
		prefix = "udp:"
	case Network_UNIX:
		prefix = "unix:"
	}
	return prefix + d.NetAddr()
}

func (d Destination) IsValid() bool {
	return d.Network != Network_Unknown
}

func (p *Endpoint) AsDestination() Destination {
	return Destination{
		Network: p.Network,
		Address: p.Address.AsAddress(),
		Port:    Port(p.Port),
	}
}
