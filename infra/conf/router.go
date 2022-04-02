package conf

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/vmessocket/vmessocket/app/router"
	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/infra/conf/cfgcommon"
	"github.com/vmessocket/vmessocket/infra/conf/geodata"
)

type BalancingRule struct {
	Tag       string               `json:"tag"`
	Selectors cfgcommon.StringList `json:"selector"`
	Strategy  StrategyConfig       `json:"strategy"`
}

type RouterConfig struct {
	Settings       *RouterRulesConfig `json:"settings"`
	RuleList       []json.RawMessage  `json:"rules"`
	Balancers      []*BalancingRule   `json:"balancers"`
	DomainMatcher string `json:"domainMatcher"`
}

type RouterRulesConfig struct {
	RuleList       []json.RawMessage `json:"rules"`
}

func (r *BalancingRule) Build() (*router.BalancingRule, error) {
	if r.Tag == "" {
		return nil, newError("empty balancer tag")
	}
	if len(r.Selectors) == 0 {
		return nil, newError("empty selector list")
	}
	var strategy string
	switch strings.ToLower(r.Strategy.Type) {
	case strategyRandom, "":
		strategy = strategyRandom
	case strategyLeastPing:
		strategy = "leastPing"
	default:
		return nil, newError("unknown balancing strategy: " + r.Strategy.Type)
	}
	return &router.BalancingRule{
		Tag:              r.Tag,
		OutboundSelector: []string(r.Selectors),
		Strategy:         strategy,
	}, nil
}

func (c *RouterConfig) Build() (*router.Config, error) {
	config := new(router.Config)
	config.DomainStrategy = c.getDomainStrategy()
	cfgctx := cfgcommon.NewConfigureLoadingContext(context.Background())
	geoloadername := platform.NewEnvFlag("vmessocket.conf.geoloader").GetValue(func() string {
		return "standard"
	})
	if loader, err := geodata.GetGeoDataLoader(geoloadername); err == nil {
		cfgcommon.SetGeoDataLoader(cfgctx, loader)
	} else {
		return nil, newError("unable to create geo data loader ").Base(err)
	}
	for _, rawBalancer := range c.Balancers {
		balancer, err := rawBalancer.Build()
		if err != nil {
			return nil, err
		}
		config.BalancingRule = append(config.BalancingRule, balancer)
	}
	return config, nil
}
