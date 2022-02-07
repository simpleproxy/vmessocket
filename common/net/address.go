package net

import (
	"bytes"
	"net"
	"strings"
)

const (
	AddressFamilyIPv4   = AddressFamily(0)
	AddressFamilyIPv6   = AddressFamily(1)
	AddressFamilyDomain = AddressFamily(2)
)

var (
	LocalHostIP     = IPAddress([]byte{127, 0, 0, 1})
	AnyIP           = IPAddress([]byte{0, 0, 0, 0})
	LocalHostDomain = DomainAddress("localhost")
	LocalHostIPv6   = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	AnyIPv6         = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	bytes0 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type Address interface {
	IP() net.IP
	Domain() string
	Family() AddressFamily
	String() string
}

type AddressFamily byte

type domainAddress string

type ipv4Address [4]byte

type ipv6Address [16]byte

func DomainAddress(domain string) Address {
	return domainAddress(domain)
}

func IPAddress(ip []byte) Address {
	switch len(ip) {
	case net.IPv4len:
		var addr ipv4Address = [4]byte{ip[0], ip[1], ip[2], ip[3]}
		return addr
	case net.IPv6len:
		if bytes.Equal(ip[:10], bytes0) && ip[10] == 0xff && ip[11] == 0xff {
			return IPAddress(ip[12:16])
		}
		var addr ipv6Address = [16]byte{
			ip[0], ip[1], ip[2], ip[3],
			ip[4], ip[5], ip[6], ip[7],
			ip[8], ip[9], ip[10], ip[11],
			ip[12], ip[13], ip[14], ip[15],
		}
		return addr
	default:
		newError("invalid IP format: ", ip).AtError().WriteToLog()
		return nil
	}
}

func isAlphaNum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func NewIPOrDomain(addr Address) *IPOrDomain {
	switch addr.Family() {
	case AddressFamilyDomain:
		return &IPOrDomain{
			Address: &IPOrDomain_Domain{
				Domain: addr.Domain(),
			},
		}
	case AddressFamilyIPv4, AddressFamilyIPv6:
		return &IPOrDomain{
			Address: &IPOrDomain_Ip{
				Ip: addr.IP(),
			},
		}
	default:
		panic("Unknown Address type.")
	}
}

func ParseAddress(addr string) Address {
	lenAddr := len(addr)
	if lenAddr > 0 && addr[0] == '[' && addr[lenAddr-1] == ']' {
		addr = addr[1 : lenAddr-1]
		lenAddr -= 2
	}
	if lenAddr > 0 && (!isAlphaNum(addr[0]) || !isAlphaNum(addr[len(addr)-1])) {
		addr = strings.TrimSpace(addr)
	}
	ip := net.ParseIP(addr)
	if ip != nil {
		return IPAddress(ip)
	}
	return DomainAddress(addr)
}

func (d *IPOrDomain) AsAddress() Address {
	if d == nil {
		return nil
	}
	switch addr := d.Address.(type) {
	case *IPOrDomain_Ip:
		return IPAddress(addr.Ip)
	case *IPOrDomain_Domain:
		return DomainAddress(addr.Domain)
	}
	panic("Common|Net: Invalid address.")
}

func (a domainAddress) Domain() string {
	return string(a)
}

func (ipv4Address) Domain() string {
	panic("Calling Domain() on an IPv4Address.")
}

func (ipv6Address) Domain() string {
	panic("Calling Domain() on an IPv6Address.")
}

func (domainAddress) Family() AddressFamily {
	return AddressFamilyDomain
}

func (ipv4Address) Family() AddressFamily {
	return AddressFamilyIPv4
}

func (ipv6Address) Family() AddressFamily {
	return AddressFamilyIPv6
}

func (domainAddress) IP() net.IP {
	panic("Calling IP() on a DomainAddress.")
}

func (a ipv4Address) IP() net.IP {
	return net.IP(a[:])
}

func (a ipv6Address) IP() net.IP {
	return net.IP(a[:])
}

func (af AddressFamily) IsDomain() bool {
	return af == AddressFamilyDomain
}

func (af AddressFamily) IsIP() bool {
	return af == AddressFamilyIPv4 || af == AddressFamilyIPv6
}

func (af AddressFamily) IsIPv4() bool {
	return af == AddressFamilyIPv4
}

func (af AddressFamily) IsIPv6() bool {
	return af == AddressFamilyIPv6
}

func (a domainAddress) String() string {
	return a.Domain()
}

func (a ipv4Address) String() string {
	return a.IP().String()
}

func (a ipv6Address) String() string {
	return "[" + a.IP().String() + "]"
}
