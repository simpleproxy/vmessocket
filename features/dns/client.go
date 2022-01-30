package dns

import (
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/features"
)

type IPOption struct {
	IPv4Enable bool
	IPv6Enable bool
	FakeEnable bool
}

type Client interface {
	features.Feature
	LookupIP(domain string) ([]net.IP, error)
}

type IPv4Lookup interface {
	LookupIPv4(domain string) ([]net.IP, error)
}

type IPv6Lookup interface {
	LookupIPv6(domain string) ([]net.IP, error)
}

type ClientWithIPOption interface {
	GetIPOption() *IPOption
	SetQueryOption(isIPv4Enable, isIPv6Enable bool)
	SetFakeDNSOption(isFakeEnable bool)
}

func ClientType() interface{} {
	return (*Client)(nil)
}

var ErrEmptyResponse = errors.New("empty response")

type RCodeError uint16

func (e RCodeError) Error() string {
	return serial.Concat("rcode: ", uint16(e))
}

func RCodeFromError(err error) uint16 {
	if err == nil {
		return 0
	}
	cause := errors.Cause(err)
	if r, ok := cause.(RCodeError); ok {
		return uint16(r)
	}
	return 0
}
