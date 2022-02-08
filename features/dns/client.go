package dns

import (
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/features"
)

var ErrEmptyResponse = errors.New("empty response")

type Client interface {
	features.Feature
	LookupIP(domain string) ([]net.IP, error)
}

type ClientWithIPOption interface {
	GetIPOption() *IPOption
	SetQueryOption(isIPv4Enable, isIPv6Enable bool)
}

type IPOption struct {
	IPv4Enable bool
	IPv6Enable bool
}

type IPv4Lookup interface {
	LookupIPv4(domain string) ([]net.IP, error)
}

type IPv6Lookup interface {
	LookupIPv6(domain string) ([]net.IP, error)
}

type RCodeError uint16

func ClientType() interface{} {
	return (*Client)(nil)
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

func (e RCodeError) Error() string {
	return serial.Concat("rcode: ", uint16(e))
}
