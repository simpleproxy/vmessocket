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

type RouterRulesConfig struct {
	RuleList       []json.RawMessage `json:"rules"`
	DomainStrategy string            `json:"domainStrategy"`
}

type StrategyConfig struct {
	Type     string           `json:"type"`
	Settings *json.RawMessage `json:"settings"`
}

type BalancingRule struct {
	Tag       string               `json:"tag"`
	Selectors cfgcommon.StringList `json:"selector"`
	Strategy  StrategyConfig       `json:"strategy"`
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

type RouterConfig struct {
	Settings       *RouterRulesConfig `json:"settings"` // Deprecated
	RuleList       []json.RawMessage  `json:"rules"`
	DomainStrategy *string            `json:"domainStrategy"`
	Balancers      []*BalancingRule   `json:"balancers"`

	DomainMatcher string `json:"domainMatcher"`
}

func (c *RouterConfig) getDomainStrategy() router.Config_DomainStrategy {
	ds := ""
	if c.DomainStrategy != nil {
		ds = *c.DomainStrategy
	} else if c.Settings != nil {
		ds = c.Settings.DomainStrategy
	}

	switch strings.ToLower(ds) {
	case "alwaysip", "always_ip", "always-ip":
		return router.Config_UseIp
	case "ipifnonmatch", "ip_if_non_match", "ip-if-non-match":
		return router.Config_IpIfNonMatch
	case "ipondemand", "ip_on_demand", "ip-on-demand":
		return router.Config_IpOnDemand
	default:
		return router.Config_AsIs
	}
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
