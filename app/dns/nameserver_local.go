//go:build !confonly
// +build !confonly

package dns

import (
	"context"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/dns/localdns"
)

type LocalNameServer struct {
	client *localdns.Client
}

func (s *LocalNameServer) QueryIP(_ context.Context, domain string, _ net.IP, option dns.IPOption, _ bool) ([]net.IP, error) {
	var ips []net.IP
	var err error

	switch {
	case option.IPv4Enable && option.IPv6Enable:
		ips, err = s.client.LookupIP(domain)
	case option.IPv4Enable:
		ips, err = s.client.LookupIPv4(domain)
	case option.IPv6Enable:
		ips, err = s.client.LookupIPv6(domain)
	}

	if len(ips) > 0 {
		newError("Localhost got answer: ", domain, " -> ", ips).AtInfo().WriteToLog()
	}

	return ips, err
}

func (s *LocalNameServer) Name() string {
	return "localhost"
}

func NewLocalNameServer() *LocalNameServer {
	newError("DNS: created localhost client").AtInfo().WriteToLog()
	return &LocalNameServer{
		client: localdns.New(),
	}
}

func NewLocalDNSClient() *Client {
	return &Client{server: NewLocalNameServer()}
}
