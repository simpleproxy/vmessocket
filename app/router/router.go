package router

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/outbound"
	"github.com/vmessocket/vmessocket/features/routing"
)

type Route struct {
	routing.Context
	outboundGroupTags []string
	outboundTag       string
}

type Router struct {
	rules          []*Rule
	balancers      map[string]*Balancer
	dns            dns.Client
}

func (*Router) Close() error {
	return nil
}

func (r *Route) GetOutboundGroupTags() []string {
	return r.outboundGroupTags
}

func (r *Route) GetOutboundTag() string {
	return r.outboundTag
}

func (r *Router) Init(ctx context.Context, config *Config, d dns.Client, ohm outbound.Manager) error {
	r.dns = d
	r.balancers = make(map[string]*Balancer, len(config.BalancingRule))
	for _, rule := range config.BalancingRule {
		balancer, err := rule.Build(ohm)
		if err != nil {
			return err
		}
		r.balancers[rule.Tag] = balancer
	}
	r.rules = make([]*Rule, 0, len(config.Rule))
	for _, rule := range config.Rule {
		cond, err := rule.BuildCondition()
		if err != nil {
			return err
		}
		rr := &Rule{
			Condition: cond,
			Tag:       rule.GetTag(),
		}
		btag := rule.GetBalancingTag()
		if len(btag) > 0 {
			brule, found := r.balancers[btag]
			if !found {
				return newError("balancer ", btag, " not found")
			}
			rr.Balancer = brule
		}
		r.rules = append(r.rules, rr)
	}
	return nil
}

func (r *Router) PickRoute(ctx routing.Context) (routing.Route, error) {
	rule, ctx, err := r.pickRouteInternal(ctx)
	if err != nil {
		return nil, err
	}
	tag, err := rule.GetTag()
	if err != nil {
		return nil, err
	}
	return &Route{Context: ctx, outboundTag: tag}, nil
}

func (r *Router) pickRouteInternal(ctx routing.Context) (*Rule, routing.Context, error) {
	skipDNSResolve := ctx.GetSkipDNSResolve()
	for _, rule := range r.rules {
		if rule.Apply(ctx) {
			return rule, ctx, nil
		}
	}
	for _, rule := range r.rules {
		if rule.Apply(ctx) {
			return rule, ctx, nil
		}
	}
	return nil, ctx, common.ErrNoClue
}

func (*Router) Start() error {
	return nil
}

func (*Router) Type() interface{} {
	return routing.RouterType()
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		r := new(Router)
		if err := core.RequireFeatures(ctx, func(d dns.Client, ohm outbound.Manager) error {
			return r.Init(ctx, config.(*Config), d, ohm)
		}); err != nil {
			return nil, err
		}
		return r, nil
	}))
}
