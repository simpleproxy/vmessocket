package localdns

import (
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features/dns"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (*Client) Close() error {
	return nil
}

func (*Client) LookupIP(host string) ([]net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	parsedIPs := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		parsed := net.IPAddress(ip)
		if parsed != nil {
			parsedIPs = append(parsedIPs, parsed.IP())
		}
	}
	if len(parsedIPs) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return parsedIPs, nil
}

func (c *Client) LookupIPv4(host string) ([]net.IP, error) {
	ips, err := c.LookupIP(host)
	if err != nil {
		return nil, err
	}
	ipv4 := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if len(ip) == net.IPv4len {
			ipv4 = append(ipv4, ip)
		}
	}
	if len(ipv4) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return ipv4, nil
}

func (c *Client) LookupIPv6(host string) ([]net.IP, error) {
	ips, err := c.LookupIP(host)
	if err != nil {
		return nil, err
	}
	ipv6 := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if len(ip) == net.IPv6len {
			ipv6 = append(ipv6, ip)
		}
	}
	if len(ipv6) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return ipv6, nil
}

func (*Client) Start() error {
	return nil
}

func (*Client) Type() interface{} {
	return dns.ClientType()
}
