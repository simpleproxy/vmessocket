package dns

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/vmessocket/vmessocket/app/router"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/strmatcher"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/routing"
)

var errExpectedIPNonMatch = errors.New("expectIPs not match")

type Client struct {
	server       Server
	clientIP     net.IP
	skipFallback bool
	domains      []string
	expectIPs    []*router.GeoIPMatcher
}

type Server interface {
	Name() string
	QueryIP(ctx context.Context, domain string, clientIP net.IP, option dns.IPOption, disableCache bool) ([]net.IP, error)
}

func NewClient(ctx context.Context, ns *NameServer, clientIP net.IP, container router.GeoIPMatcherContainer, matcherInfos *[]*DomainMatcherInfo, updateDomainRule func(strmatcher.Matcher, int, []*DomainMatcherInfo) error) (*Client, error) {
	client := &Client{}

	err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error {
		server, err := NewServer(ns.Address.AsDestination(), dispatcher)
		if err != nil {
			return newError("failed to create nameserver").Base(err).AtWarning()
		}

		if _, isLocalDNS := server.(*LocalNameServer); isLocalDNS {
			ns.PrioritizedDomain = append(ns.PrioritizedDomain, localTLDsAndDotlessDomains...)
			ns.OriginalRules = append(ns.OriginalRules, localTLDsAndDotlessDomainsRule)
			for i := 0; i < len(localTLDsAndDotlessDomains); i++ {
				*matcherInfos = append(*matcherInfos, &DomainMatcherInfo{
					clientIdx:     uint16(0),
					domainRuleIdx: uint16(0),
				})
			}
		}

		var rules []string
		ruleCurr := 0
		ruleIter := 0
		for _, domain := range ns.PrioritizedDomain {
			domainRule, err := toStrMatcher(domain.Type, domain.Domain)
			if err != nil {
				return newError("failed to create prioritized domain").Base(err).AtWarning()
			}
			originalRuleIdx := ruleCurr
			if ruleCurr < len(ns.OriginalRules) {
				rule := ns.OriginalRules[ruleCurr]
				if ruleCurr >= len(rules) {
					rules = append(rules, rule.Rule)
				}
				ruleIter++
				if ruleIter >= int(rule.Size) {
					ruleIter = 0
					ruleCurr++
				}
			} else {
				rules = append(rules, domainRule.String())
				ruleCurr++
			}
			err = updateDomainRule(domainRule, originalRuleIdx, *matcherInfos)
			if err != nil {
				return newError("failed to create prioritized domain").Base(err).AtWarning()
			}
		}

		var matchers []*router.GeoIPMatcher
		for _, geoip := range ns.Geoip {
			matcher, err := container.Add(geoip)
			if err != nil {
				return newError("failed to create ip matcher").Base(err).AtWarning()
			}
			matchers = append(matchers, matcher)
		}

		if len(clientIP) > 0 {
			switch ns.Address.Address.GetAddress().(type) {
			case *net.IPOrDomain_Domain:
				newError("DNS: client ", ns.Address.Address.GetDomain(), " uses clientIP ", clientIP.String()).AtInfo().WriteToLog()
			case *net.IPOrDomain_Ip:
				newError("DNS: client ", ns.Address.Address.GetIp(), " uses clientIP ", clientIP.String()).AtInfo().WriteToLog()
			}
		}

		client.server = server
		client.clientIP = clientIP
		client.skipFallback = ns.SkipFallback
		client.domains = rules
		client.expectIPs = matchers
		return nil
	})
	return client, err
}

func NewServer(dest net.Destination, dispatcher routing.Dispatcher) (Server, error) {
	if address := dest.Address; address.Family().IsDomain() {
		u, err := url.Parse(address.Domain())
		if err != nil {
			return nil, err
		}
		switch {
		case strings.EqualFold(u.String(), "localhost"):
			return NewLocalNameServer(), nil
		case strings.EqualFold(u.Scheme, "https"):
			return NewDoHNameServer(u, dispatcher)
		case strings.EqualFold(u.Scheme, "https+local"):
			return NewDoHLocalNameServer(u), nil
		case strings.EqualFold(u.Scheme, "tcp"):
			return NewTCPNameServer(u, dispatcher)
		case strings.EqualFold(u.Scheme, "tcp+local"):
			return NewTCPLocalNameServer(u)
		}
	}
	if dest.Network == net.Network_Unknown {
		dest.Network = net.Network_UDP
	}
	if dest.Network == net.Network_UDP {
		return NewClassicNameServer(dest, dispatcher), nil
	}
	return nil, newError("No available name server could be created from ", dest).AtWarning()
}

func NewSimpleClient(ctx context.Context, endpoint *net.Endpoint, clientIP net.IP) (*Client, error) {
	client := &Client{}
	err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error {
		server, err := NewServer(endpoint.AsDestination(), dispatcher)
		if err != nil {
			return newError("failed to create nameserver").Base(err).AtWarning()
		}
		client.server = server
		client.clientIP = clientIP
		return nil
	})

	if len(clientIP) > 0 {
		switch endpoint.Address.GetAddress().(type) {
		case *net.IPOrDomain_Domain:
			newError("DNS: client ", endpoint.Address.GetDomain(), " uses clientIP ", clientIP.String()).AtInfo().WriteToLog()
		case *net.IPOrDomain_Ip:
			newError("DNS: client ", endpoint.Address.GetIp(), " uses clientIP ", clientIP.String()).AtInfo().WriteToLog()
		}
	}

	return client, err
}

func (c *Client) MatchExpectedIPs(domain string, ips []net.IP) ([]net.IP, error) {
	if len(c.expectIPs) == 0 {
		return ips, nil
	}
	newIps := []net.IP{}
	for _, ip := range ips {
		for _, matcher := range c.expectIPs {
			if matcher.Match(ip) {
				newIps = append(newIps, ip)
				break
			}
		}
	}
	if len(newIps) == 0 {
		return nil, errExpectedIPNonMatch
	}
	newError("domain ", domain, " expectIPs ", newIps, " matched at server ", c.Name()).AtDebug().WriteToLog()
	return newIps, nil
}

func (c *Client) Name() string {
	return c.server.Name()
}

func (c *Client) QueryIP(ctx context.Context, domain string, option dns.IPOption, disableCache bool) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	ips, err := c.server.QueryIP(ctx, domain, c.clientIP, option, disableCache)
	cancel()

	if err != nil {
		return ips, err
	}
	return c.MatchExpectedIPs(domain, ips)
}
