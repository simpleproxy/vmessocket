package geodata

import (
	"strings"

	"github.com/vmessocket/vmessocket/app/router"
)

type AttributeList struct {
	matcher []AttributeMatcher
}

type AttributeMatcher interface {
	Match(*router.Domain) bool
}

type BooleanMatcher string

func (al *AttributeList) Match(domain *router.Domain) bool {
	for _, matcher := range al.matcher {
		if !matcher.Match(domain) {
			return false
		}
	}
	return true
}

func parseAttrs(attrs []string) *AttributeList {
	al := new(AttributeList)
	for _, attr := range attrs {
		trimmedAttr := strings.ToLower(strings.TrimSpace(attr))
		if len(trimmedAttr) == 0 {
			continue
		}
		al.matcher = append(al.matcher, BooleanMatcher(trimmedAttr))
	}
	return al
}

func (al *AttributeList) IsEmpty() bool {
	return len(al.matcher) == 0
}

func (m BooleanMatcher) Match(domain *router.Domain) bool {
	for _, attr := range domain.Attribute {
		if strings.EqualFold(attr.GetKey(), string(m)) {
			return true
		}
	}
	return false
}
