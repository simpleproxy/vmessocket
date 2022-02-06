package router

import (
	"github.com/vmessocket/vmessocket/common/dice"
	"github.com/vmessocket/vmessocket/features/outbound"
)

type Balancer struct {
	selectors []string
	strategy  BalancingStrategy
	ohm       outbound.Manager
}

type BalancingStrategy interface {
	PickOutbound([]string) string
}

type RandomStrategy struct{}

func (b *Balancer) PickOutbound() (string, error) {
	hs, ok := b.ohm.(outbound.HandlerSelector)
	if !ok {
		return "", newError("outbound.Manager is not a HandlerSelector")
	}
	tags := hs.Select(b.selectors)
	if len(tags) == 0 {
		return "", newError("no available outbounds selected")
	}
	tag := b.strategy.PickOutbound(tags)
	if tag == "" {
		return "", newError("balancing strategy returns empty tag")
	}
	return tag, nil
}

func (s *RandomStrategy) PickOutbound(tags []string) string {
	n := len(tags)
	if n == 0 {
		panic("0 tags")
	}
	return tags[dice.Roll(n)]
}
