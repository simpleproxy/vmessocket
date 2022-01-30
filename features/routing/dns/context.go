package dns

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/routing"
)

type ResolvableContext struct {
	routing.Context
	dnsClient   dns.Client
	resolvedIPs []net.IP
}

func (ctx *ResolvableContext) GetTargetIPs() []net.IP {
	if ips := ctx.Context.GetTargetIPs(); len(ips) != 0 {
		return ips
	}

	if len(ctx.resolvedIPs) > 0 {
		return ctx.resolvedIPs
	}

	if domain := ctx.GetTargetDomain(); len(domain) != 0 {
		lookupFunc := ctx.dnsClient.LookupIP
		ipOption := &dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: true,
		}

		if c, ok := ctx.dnsClient.(dns.ClientWithIPOption); ok {
			ipOption = c.GetIPOption()
			c.SetFakeDNSOption(false)
		} else {
			newError("ctx.dnsClient doesn't implement ClientWithIPOption").AtDebug().WriteToLog()
		}

		switch {
		case ipOption.IPv4Enable && !ipOption.IPv6Enable:
			if lookupIPv4, ok := ctx.dnsClient.(dns.IPv4Lookup); ok {
				lookupFunc = lookupIPv4.LookupIPv4
			} else {
				newError("ctx.dnsClient doesn't implement IPv4Lookup. Use LookupIP instead.").AtDebug().WriteToLog()
			}
		case !ipOption.IPv4Enable && ipOption.IPv6Enable:
			if lookupIPv6, ok := ctx.dnsClient.(dns.IPv6Lookup); ok {
				lookupFunc = lookupIPv6.LookupIPv6
			} else {
				newError("ctx.dnsClient doesn't implement IPv6Lookup. Use LookupIP instead.").AtDebug().WriteToLog()
			}
		}

		ips, err := lookupFunc(domain)
		if err == nil {
			ctx.resolvedIPs = ips
			return ips
		}
		newError("resolve ip for ", domain).Base(err).WriteToLog()
	}

	return nil
}

func ContextWithDNSClient(ctx routing.Context, client dns.Client) routing.Context {
	return &ResolvableContext{Context: ctx, dnsClient: client}
}
