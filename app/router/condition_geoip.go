//go:build !confonly
// +build !confonly

package router

import (
	"inet.af/netaddr"

	"github.com/vmessocket/vmessocket/common/net"
)

type GeoIPMatcher struct {
	countryCode  string
	reverseMatch bool
	ip4          *netaddr.IPSet
	ip6          *netaddr.IPSet
}

func (m *GeoIPMatcher) Init(cidrs []*CIDR) error {
	var builder4, builder6 netaddr.IPSetBuilder
	for _, cidr := range cidrs {
		netaddrIP, ok := netaddr.FromStdIP(net.IP(cidr.GetIp()))
		if !ok {
			return newError("invalid IP address ", cidr)
		}
		ipPrefix := netaddr.IPPrefixFrom(netaddrIP, uint8(cidr.GetPrefix()))
		switch {
		case netaddrIP.Is4():
			builder4.AddPrefix(ipPrefix)
		case netaddrIP.Is6():
			builder6.AddPrefix(ipPrefix)
		}
	}

	var err error
	m.ip4, err = builder4.IPSet()
	if err != nil {
		return err
	}
	m.ip6, err = builder6.IPSet()
	if err != nil {
		return err
	}

	return nil
}

func (m *GeoIPMatcher) SetReverseMatch(isReverseMatch bool) {
	m.reverseMatch = isReverseMatch
}

func (m *GeoIPMatcher) match4(ip net.IP) bool {
	nip, ok := netaddr.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ip4.Contains(nip)
}

func (m *GeoIPMatcher) match6(ip net.IP) bool {
	nip, ok := netaddr.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ip6.Contains(nip)
}

func (m *GeoIPMatcher) Match(ip net.IP) bool {
	isMatched := false
	switch len(ip) {
	case net.IPv4len:
		isMatched = m.match4(ip)
	case net.IPv6len:
		isMatched = m.match6(ip)
	}
	if m.reverseMatch {
		return !isMatched
	}
	return isMatched
}

type GeoIPMatcherContainer struct {
	matchers []*GeoIPMatcher
}

func (c *GeoIPMatcherContainer) Add(geoip *GeoIP) (*GeoIPMatcher, error) {
	if geoip.CountryCode != "" {
		for _, m := range c.matchers {
			if m.countryCode == geoip.CountryCode && m.reverseMatch == geoip.ReverseMatch {
				return m, nil
			}
		}
	}

	m := &GeoIPMatcher{
		countryCode:  geoip.CountryCode,
		reverseMatch: geoip.ReverseMatch,
	}
	if err := m.Init(geoip.Cidr); err != nil {
		return nil, err
	}
	if geoip.CountryCode != "" {
		c.matchers = append(c.matchers, m)
	}
	return m, nil
}

var globalGeoIPContainer GeoIPMatcherContainer
