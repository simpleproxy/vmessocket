package strmatcher

import (
	"regexp"
)

type Matcher interface {
	Match(string) bool
	String() string
}

type Type byte

const (
	Full Type = iota
	Substr
	Domain
	Regex
)

func (t Type) New(pattern string) (Matcher, error) {
	switch t {
	case Full:
		return fullMatcher(pattern), nil
	case Substr:
		return substrMatcher(pattern), nil
	case Domain:
		return domainMatcher(pattern), nil
	case Regex:
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return &regexMatcher{
			pattern: r,
		}, nil
	default:
		panic("Unknown type")
	}
}

type IndexMatcher interface {
	Match(input string) []uint32
}

type matcherEntry struct {
	m  Matcher
	id uint32
}

type MatcherGroup struct {
	count         uint32
	fullMatcher   FullMatcherGroup
	domainMatcher DomainMatcherGroup
	otherMatchers []matcherEntry
}

func (g *MatcherGroup) Add(m Matcher) uint32 {
	g.count++
	c := g.count

	switch tm := m.(type) {
	case fullMatcher:
		g.fullMatcher.addMatcher(tm, c)
	case domainMatcher:
		g.domainMatcher.addMatcher(tm, c)
	default:
		g.otherMatchers = append(g.otherMatchers, matcherEntry{
			m:  m,
			id: c,
		})
	}

	return c
}

func (g *MatcherGroup) Match(pattern string) []uint32 {
	result := []uint32{}
	result = append(result, g.fullMatcher.Match(pattern)...)
	result = append(result, g.domainMatcher.Match(pattern)...)
	for _, e := range g.otherMatchers {
		if e.m.Match(pattern) {
			result = append(result, e.id)
		}
	}
	return result
}

func (g *MatcherGroup) Size() uint32 {
	return g.count
}
