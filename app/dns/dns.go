package dns

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/vmessocket/vmessocket/app/router"
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/common/strmatcher"
	"github.com/vmessocket/vmessocket/features"
	"github.com/vmessocket/vmessocket/features/dns"
)

type DomainMatcherInfo struct {
	clientIdx     uint16
	domainRuleIdx uint16
}

type DNS struct {
	sync.Mutex
	disableCache           bool
	disableFallback        bool
	disableFallbackIfMatch bool
	ipOption               *dns.IPOption
	hosts                  *StaticHosts
	clients                []*Client
	ctx                    context.Context
	domainMatcher          strmatcher.IndexMatcher
	matcherInfos           []*DomainMatcherInfo
}

func New(ctx context.Context, config *Config) (*DNS, error) {
	var clientIP net.IP
	switch len(config.ClientIp) {
	case 0, net.IPv4len, net.IPv6len:
		clientIP = net.IP(config.ClientIp)
	default:
		return nil, newError("unexpected client IP length ", len(config.ClientIp))
	}
	var ipOption *dns.IPOption
	switch config.QueryStrategy {
	case QueryStrategy_USE_IP:
		ipOption = &dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: true,
		}
	case QueryStrategy_USE_IP4:
		ipOption = &dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: false,
		}
	case QueryStrategy_USE_IP6:
		ipOption = &dns.IPOption{
			IPv4Enable: false,
			IPv6Enable: true,
		}
	}
	hosts, err := NewStaticHosts(config.StaticHosts, config.Hosts)
	if err != nil {
		return nil, newError("failed to create hosts").Base(err)
	}
	clients := []*Client{}
	domainRuleCount := 0
	for _, ns := range config.NameServer {
		domainRuleCount += len(ns.PrioritizedDomain)
	}
	matcherInfos := make([]*DomainMatcherInfo, domainRuleCount+1)
	domainMatcher := &strmatcher.MatcherGroup{}
	geoipContainer := router.GeoIPMatcherContainer{}
	for _, endpoint := range config.NameServers {
		features.PrintDeprecatedFeatureWarning("simple DNS server")
		client, err := NewSimpleClient(ctx, endpoint, clientIP)
		if err != nil {
			return nil, newError("failed to create client").Base(err)
		}
		clients = append(clients, client)
	}
	for _, ns := range config.NameServer {
		clientIdx := len(clients)
		updateDomain := func(domainRule strmatcher.Matcher, originalRuleIdx int, matcherInfos []*DomainMatcherInfo) error {
			midx := domainMatcher.Add(domainRule)
			matcherInfos[midx] = &DomainMatcherInfo{
				clientIdx:     uint16(clientIdx),
				domainRuleIdx: uint16(originalRuleIdx),
			}
			return nil
		}
		myClientIP := clientIP
		switch len(ns.ClientIp) {
		case net.IPv4len, net.IPv6len:
			myClientIP = net.IP(ns.ClientIp)
		}
		client, err := NewClient(ctx, ns, myClientIP, geoipContainer, &matcherInfos, updateDomain)
		if err != nil {
			return nil, newError("failed to create client").Base(err)
		}
		clients = append(clients, client)
	}
	return &DNS{
		hosts:                  hosts,
		ipOption:               ipOption,
		clients:                clients,
		ctx:                    ctx,
		domainMatcher:          domainMatcher,
		matcherInfos:           matcherInfos,
		disableCache:           config.DisableCache,
		disableFallback:        config.DisableFallback,
		disableFallbackIfMatch: config.DisableFallbackIfMatch,
	}, nil
}

func (s *DNS) Close() error {
	return nil
}

func (s *DNS) GetIPOption() *dns.IPOption {
	return s.ipOption
}

func (s *DNS) IsOwnLink(ctx context.Context) bool {
	inbound := session.InboundFromContext(ctx)
	return inbound != nil
}

func (s *DNS) LookupIP(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, *s.ipOption)
}

func (s *DNS) lookupIPInternal(domain string, option dns.IPOption) ([]net.IP, error) {
	if domain == "" {
		return nil, newError("empty domain name")
	}
	domain = strings.TrimSuffix(domain, ".")
	switch addrs := s.hosts.Lookup(domain, option); {
	case addrs == nil:
		break
	case len(addrs) == 0:
		return nil, dns.ErrEmptyResponse
	case len(addrs) == 1 && addrs[0].Family().IsDomain():
		newError("domain replaced: ", domain, " -> ", addrs[0].Domain()).WriteToLog()
		domain = addrs[0].Domain()
	default:
		newError("returning ", len(addrs), " IP(s) for domain ", domain, " -> ", addrs).WriteToLog()
		return toNetIP(addrs)
	}
	errs := []error{}
	ctx := session.ContextWithInbound(s.ctx, &session.Inbound{})
	for _, client := range s.sortClients(domain) {
		ips, err := client.QueryIP(ctx, domain, option, s.disableCache)
		if len(ips) > 0 {
			return ips, nil
		}
		if err != nil {
			newError("failed to lookup ip for domain ", domain, " at server ", client.Name()).Base(err).WriteToLog()
			errs = append(errs, err)
		}
		if err != context.Canceled && err != context.DeadlineExceeded && err != errExpectedIPNonMatch {
			return nil, err
		}
	}
	return nil, newError("returning nil for domain ", domain).Base(errors.Combine(errs...))
}

func (s *DNS) LookupIPv4(domain string) ([]net.IP, error) {
	if !s.ipOption.IPv4Enable {
		return nil, dns.ErrEmptyResponse
	}
	o := *s.ipOption
	o.IPv6Enable = false
	return s.lookupIPInternal(domain, o)
}

func (s *DNS) LookupIPv6(domain string) ([]net.IP, error) {
	if !s.ipOption.IPv6Enable {
		return nil, dns.ErrEmptyResponse
	}
	o := *s.ipOption
	o.IPv4Enable = false
	return s.lookupIPInternal(domain, o)
}

func (s *DNS) SetQueryOption(isIPv4Enable, isIPv6Enable bool) {
	s.ipOption.IPv4Enable = isIPv4Enable
	s.ipOption.IPv6Enable = isIPv6Enable
}

func (s *DNS) sortClients(domain string) []*Client {
	clients := make([]*Client, 0, len(s.clients))
	clientUsed := make([]bool, len(s.clients))
	clientNames := make([]string, 0, len(s.clients))
	domainRules := []string{}
	hasMatch := false
	for _, match := range s.domainMatcher.Match(domain) {
		info := s.matcherInfos[match]
		client := s.clients[info.clientIdx]
		domainRule := client.domains[info.domainRuleIdx]
		domainRules = append(domainRules, fmt.Sprintf("%s(DNS idx:%d)", domainRule, info.clientIdx))
		if clientUsed[info.clientIdx] {
			continue
		}
		clientUsed[info.clientIdx] = true
		clients = append(clients, client)
		clientNames = append(clientNames, client.Name())
		hasMatch = true
	}
	if !(s.disableFallback || s.disableFallbackIfMatch && hasMatch) {
		for idx, client := range s.clients {
			if clientUsed[idx] || client.skipFallback {
				continue
			}
			clientUsed[idx] = true
			clients = append(clients, client)
			clientNames = append(clientNames, client.Name())
		}
	}
	if len(domainRules) > 0 {
		newError("domain ", domain, " matches following rules: ", domainRules).AtDebug().WriteToLog()
	}
	if len(clientNames) > 0 {
		newError("domain ", domain, " will use DNS in order: ", clientNames).AtDebug().WriteToLog()
	}
	if len(clients) == 0 {
		clients = append(clients, s.clients[0])
		clientNames = append(clientNames, s.clients[0].Name())
		newError("domain ", domain, " will use the first DNS: ", clientNames).AtDebug().WriteToLog()
	}
	return clients
}

func (s *DNS) Start() error {
	return nil
}

func (*DNS) Type() interface{} {
	return dns.ClientType()
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
