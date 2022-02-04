package conf

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/vmessocket/vmessocket/app/dispatcher"
	"github.com/vmessocket/vmessocket/app/proxyman"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/infra/conf/cfgcommon"
)

var (
	inboundConfigLoader = NewJSONConfigLoader(ConfigCreatorCache{
		"http":  func() interface{} { return new(HTTPServerConfig) },
		"socks": func() interface{} { return new(SocksServerConfig) },
		"vmess": func() interface{} { return new(VMessInboundConfig) },
	}, "protocol", "settings")

	outboundConfigLoader = NewJSONConfigLoader(ConfigCreatorCache{
		"dns":     func() interface{} { return new(DNSOutboundConfig) },
		"freedom": func() interface{} { return new(FreedomConfig) },
		"http":    func() interface{} { return new(HTTPClientConfig) },
		"socks":   func() interface{} { return new(SocksClientConfig) },
		"vmess":   func() interface{} { return new(VMessOutboundConfig) },
	}, "protocol", "settings")

	ctllog = log.New(os.Stderr, "v2ctl> ", 0)
)

func toProtocolList(s []string) ([]proxyman.KnownProtocols, error) {
	kp := make([]proxyman.KnownProtocols, 0, 8)
	for _, p := range s {
		switch strings.ToLower(p) {
		case "http":
			kp = append(kp, proxyman.KnownProtocols_HTTP)
		case "https", "tls", "ssl":
			kp = append(kp, proxyman.KnownProtocols_TLS)
		default:
			return nil, newError("Unknown protocol: ", p)
		}
	}
	return kp, nil
}

type SniffingConfig struct {
	Enabled      bool                  `json:"enabled"`
	DestOverride *cfgcommon.StringList `json:"destOverride"`
	MetadataOnly bool                  `json:"metadataOnly"`
}

func (c *SniffingConfig) Build() (*proxyman.SniffingConfig, error) {
	var p []string
	if c.DestOverride != nil {
		for _, domainOverride := range *c.DestOverride {
			switch strings.ToLower(domainOverride) {
			case "http":
				p = append(p, "http")
			case "tls", "https", "ssl":
				p = append(p, "tls")
			default:
				return nil, newError("unknown protocol: ", domainOverride)
			}
		}
	}

	return &proxyman.SniffingConfig{
		Enabled:             c.Enabled,
		DestinationOverride: p,
		MetadataOnly:        c.MetadataOnly,
	}, nil
}

type MuxConfig struct {
	Enabled     bool  `json:"enabled"`
	Concurrency int16 `json:"concurrency"`
}

func (m *MuxConfig) Build() *proxyman.MultiplexingConfig {
	if m.Concurrency < 0 {
		return nil
	}

	var con uint32 = 8
	if m.Concurrency > 0 {
		con = uint32(m.Concurrency)
	}

	return &proxyman.MultiplexingConfig{
		Enabled:     m.Enabled,
		Concurrency: con,
	}
}

type InboundDetourAllocationConfig struct {
	Strategy    string  `json:"strategy"`
	Concurrency *uint32 `json:"concurrency"`
	RefreshMin  *uint32 `json:"refresh"`
}

func (c *InboundDetourAllocationConfig) Build() (*proxyman.AllocationStrategy, error) {
	config := new(proxyman.AllocationStrategy)
	switch strings.ToLower(c.Strategy) {
	case "always":
		config.Type = proxyman.AllocationStrategy_Always
	case "random":
		config.Type = proxyman.AllocationStrategy_Random
	case "external":
		config.Type = proxyman.AllocationStrategy_External
	default:
		return nil, newError("unknown allocation strategy: ", c.Strategy)
	}
	if c.Concurrency != nil {
		config.Concurrency = &proxyman.AllocationStrategy_AllocationStrategyConcurrency{
			Value: *c.Concurrency,
		}
	}

	if c.RefreshMin != nil {
		config.Refresh = &proxyman.AllocationStrategy_AllocationStrategyRefresh{
			Value: *c.RefreshMin,
		}
	}

	return config, nil
}

type InboundDetourConfig struct {
	Protocol       string                         `json:"protocol"`
	PortRange      *cfgcommon.PortRange           `json:"port"`
	ListenOn       *cfgcommon.Address             `json:"listen"`
	Settings       *json.RawMessage               `json:"settings"`
	Tag            string                         `json:"tag"`
	Allocation     *InboundDetourAllocationConfig `json:"allocate"`
	StreamSetting  *StreamConfig                  `json:"streamSettings"`
	DomainOverride *cfgcommon.StringList          `json:"domainOverride"`
	SniffingConfig *SniffingConfig                `json:"sniffing"`
}

func (c *InboundDetourConfig) Build() (*core.InboundHandlerConfig, error) {
	receiverSettings := &proxyman.ReceiverConfig{}

	if c.ListenOn == nil {
		if c.PortRange == nil {
			return nil, newError("Listen on AnyIP but no Port(s) set in InboundDetour.")
		}
		receiverSettings.PortRange = c.PortRange.Build()
	} else {
		receiverSettings.Listen = c.ListenOn.Build()
		listenDS := c.ListenOn.Family().IsDomain() && (c.ListenOn.Domain()[0] == '/' || c.ListenOn.Domain()[0] == '@')
		listenIP := c.ListenOn.Family().IsIP() || (c.ListenOn.Family().IsDomain() && c.ListenOn.Domain() == "localhost")
		switch {
		case listenIP:
			if c.PortRange == nil {
				return nil, newError("Listen on specific ip without port in InboundDetour.")
			}
			receiverSettings.PortRange = c.PortRange.Build()
		case listenDS:
			if c.PortRange != nil {
				receiverSettings.PortRange = nil
			}
		default:
			return nil, newError("unable to listen on domain address: ", c.ListenOn.Domain())
		}
	}

	if c.Allocation != nil {
		concurrency := -1
		if c.Allocation.Concurrency != nil && c.Allocation.Strategy == "random" {
			concurrency = int(*c.Allocation.Concurrency)
		}
		portRange := int(c.PortRange.To - c.PortRange.From + 1)
		if concurrency >= 0 && concurrency >= portRange {
			return nil, newError("not enough ports. concurrency = ", concurrency, " ports: ", c.PortRange.From, " - ", c.PortRange.To)
		}

		as, err := c.Allocation.Build()
		if err != nil {
			return nil, err
		}
		receiverSettings.AllocationStrategy = as
	}
	if c.StreamSetting != nil {
		ss, err := c.StreamSetting.Build()
		if err != nil {
			return nil, err
		}
		receiverSettings.StreamSettings = ss
	}
	if c.SniffingConfig != nil {
		s, err := c.SniffingConfig.Build()
		if err != nil {
			return nil, newError("failed to build sniffing config").Base(err)
		}
		receiverSettings.SniffingSettings = s
	}
	if c.DomainOverride != nil {
		kp, err := toProtocolList(*c.DomainOverride)
		if err != nil {
			return nil, newError("failed to parse inbound detour config").Base(err)
		}
		receiverSettings.DomainOverride = kp
	}

	settings := []byte("{}")
	if c.Settings != nil {
		settings = ([]byte)(*c.Settings)
	}
	rawConfig, err := inboundConfigLoader.LoadWithID(settings, c.Protocol)
	if err != nil {
		return nil, newError("failed to load inbound detour config.").Base(err)
	}
	ts, err := rawConfig.(Buildable).Build()
	if err != nil {
		return nil, err
	}

	return &core.InboundHandlerConfig{
		Tag:              c.Tag,
		ReceiverSettings: serial.ToTypedMessage(receiverSettings),
		ProxySettings:    serial.ToTypedMessage(ts),
	}, nil
}

type OutboundDetourConfig struct {
	Protocol      string             `json:"protocol"`
	SendThrough   *cfgcommon.Address `json:"sendThrough"`
	Tag           string             `json:"tag"`
	Settings      *json.RawMessage   `json:"settings"`
	StreamSetting *StreamConfig      `json:"streamSettings"`
	ProxySettings *ProxyConfig       `json:"proxySettings"`
	MuxSettings   *MuxConfig         `json:"mux"`
}

func (c *OutboundDetourConfig) Build() (*core.OutboundHandlerConfig, error) {
	senderSettings := &proxyman.SenderConfig{}

	if c.SendThrough != nil {
		address := c.SendThrough
		if address.Family().IsDomain() {
			return nil, newError("unable to send through: " + address.String())
		}
		senderSettings.Via = address.Build()
	}

	if c.StreamSetting != nil {
		ss, err := c.StreamSetting.Build()
		if err != nil {
			return nil, err
		}
		senderSettings.StreamSettings = ss
	}

	if c.ProxySettings != nil {
		ps, err := c.ProxySettings.Build()
		if err != nil {
			return nil, newError("invalid outbound detour proxy settings.").Base(err)
		}
		senderSettings.ProxySettings = ps
	}

	if c.MuxSettings != nil {
		senderSettings.MultiplexSettings = c.MuxSettings.Build()
	}

	settings := []byte("{}")
	if c.Settings != nil {
		settings = ([]byte)(*c.Settings)
	}
	rawConfig, err := outboundConfigLoader.LoadWithID(settings, c.Protocol)
	if err != nil {
		return nil, newError("failed to parse to outbound detour config.").Base(err)
	}
	ts, err := rawConfig.(Buildable).Build()
	if err != nil {
		return nil, err
	}

	return &core.OutboundHandlerConfig{
		SenderSettings: serial.ToTypedMessage(senderSettings),
		Tag:            c.Tag,
		ProxySettings:  serial.ToTypedMessage(ts),
	}, nil
}

type Config struct {
	Port            uint16                 `json:"port"`
	InboundConfig   *InboundDetourConfig   `json:"inbound"`
	OutboundConfig  *OutboundDetourConfig  `json:"outbound"`
	InboundDetours  []InboundDetourConfig  `json:"inboundDetour"`
	OutboundDetours []OutboundDetourConfig `json:"outboundDetour"`

	LogConfig       *LogConfig             `json:"log"`
	RouterConfig    *RouterConfig          `json:"routing"`
	DNSConfig       *DNSConfig             `json:"dns"`
	InboundConfigs  []InboundDetourConfig  `json:"inbounds"`
	OutboundConfigs []OutboundDetourConfig `json:"outbounds"`
	Transport       *TransportConfig       `json:"transport"`

	Services map[string]*json.RawMessage `json:"services"`
}

func (c *Config) findInboundTag(tag string) int {
	found := -1
	for idx, ib := range c.InboundConfigs {
		if ib.Tag == tag {
			found = idx
			break
		}
	}
	return found
}

func (c *Config) findOutboundTag(tag string) int {
	found := -1
	for idx, ob := range c.OutboundConfigs {
		if ob.Tag == tag {
			found = idx
			break
		}
	}
	return found
}

func (c *Config) Override(o *Config, fn string) {
	if o.LogConfig != nil {
		c.LogConfig = o.LogConfig
	}
	if o.RouterConfig != nil {
		c.RouterConfig = o.RouterConfig
	}
	if o.DNSConfig != nil {
		c.DNSConfig = o.DNSConfig
	}
	if o.Transport != nil {
		c.Transport = o.Transport
	}
	if o.InboundConfig != nil {
		c.InboundConfig = o.InboundConfig
	}
	if o.OutboundConfig != nil {
		c.OutboundConfig = o.OutboundConfig
	}
	if o.InboundDetours != nil {
		c.InboundDetours = o.InboundDetours
	}
	if o.OutboundDetours != nil {
		c.OutboundDetours = o.OutboundDetours
	}

	if len(o.InboundConfigs) > 0 {
		if len(c.InboundConfigs) > 0 && len(o.InboundConfigs) == 1 {
			if idx := c.findInboundTag(o.InboundConfigs[0].Tag); idx > -1 {
				c.InboundConfigs[idx] = o.InboundConfigs[0]
				ctllog.Println("[", fn, "] updated inbound with tag: ", o.InboundConfigs[0].Tag)
			} else {
				c.InboundConfigs = append(c.InboundConfigs, o.InboundConfigs[0])
				ctllog.Println("[", fn, "] appended inbound with tag: ", o.InboundConfigs[0].Tag)
			}
		} else {
			c.InboundConfigs = o.InboundConfigs
		}
	}

	if len(o.OutboundConfigs) > 0 {
		if len(c.OutboundConfigs) > 0 && len(o.OutboundConfigs) == 1 {
			if idx := c.findOutboundTag(o.OutboundConfigs[0].Tag); idx > -1 {
				c.OutboundConfigs[idx] = o.OutboundConfigs[0]
				ctllog.Println("[", fn, "] updated outbound with tag: ", o.OutboundConfigs[0].Tag)
			} else {
				if strings.Contains(strings.ToLower(fn), "tail") {
					c.OutboundConfigs = append(c.OutboundConfigs, o.OutboundConfigs[0])
					ctllog.Println("[", fn, "] appended outbound with tag: ", o.OutboundConfigs[0].Tag)
				} else {
					c.OutboundConfigs = append(o.OutboundConfigs, c.OutboundConfigs...)
					ctllog.Println("[", fn, "] prepended outbound with tag: ", o.OutboundConfigs[0].Tag)
				}
			}
		} else {
			c.OutboundConfigs = o.OutboundConfigs
		}
	}
}

func applyTransportConfig(s *StreamConfig, t *TransportConfig) {
	if s.TCPSettings == nil {
		s.TCPSettings = t.TCPConfig
	}
	if s.WSSettings == nil {
		s.WSSettings = t.WSConfig
	}
	if s.HTTPSettings == nil {
		s.HTTPSettings = t.HTTPConfig
	}
}

func (c *Config) Build() (*core.Config, error) {

	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}

	var logConfMsg *serial.TypedMessage
	if c.LogConfig != nil {
		logConfMsg = serial.ToTypedMessage(c.LogConfig.Build())
	} else {
		logConfMsg = serial.ToTypedMessage(DefaultLogConfig())
	}

	config.App = append([]*serial.TypedMessage{logConfMsg}, config.App...)

	if c.RouterConfig != nil {
		routerConfig, err := c.RouterConfig.Build()
		if err != nil {
			return nil, err
		}
		config.App = append(config.App, serial.ToTypedMessage(routerConfig))
	}

	if c.DNSConfig != nil {
		dnsApp, err := c.DNSConfig.Build()
		if err != nil {
			return nil, newError("failed to parse DNS config").Base(err)
		}
		config.App = append(config.App, serial.ToTypedMessage(dnsApp))
	}

	if msg, err := c.BuildServices(c.Services); err != nil {
		return nil, newError("Cannot load service").Base(err)
	} else {
		config.App = append(config.App, msg...)
	}

	var inbounds []InboundDetourConfig

	if c.InboundConfig != nil {
		inbounds = append(inbounds, *c.InboundConfig)
	}

	if len(c.InboundDetours) > 0 {
		inbounds = append(inbounds, c.InboundDetours...)
	}

	if len(c.InboundConfigs) > 0 {
		inbounds = append(inbounds, c.InboundConfigs...)
	}

	if len(inbounds) > 0 && inbounds[0].PortRange == nil && c.Port > 0 {
		inbounds[0].PortRange = &cfgcommon.PortRange{
			From: uint32(c.Port),
			To:   uint32(c.Port),
		}
	}

	for _, rawInboundConfig := range inbounds {
		if c.Transport != nil {
			if rawInboundConfig.StreamSetting == nil {
				rawInboundConfig.StreamSetting = &StreamConfig{}
			}
			applyTransportConfig(rawInboundConfig.StreamSetting, c.Transport)
		}
		ic, err := rawInboundConfig.Build()
		if err != nil {
			return nil, err
		}
		config.Inbound = append(config.Inbound, ic)
	}

	var outbounds []OutboundDetourConfig

	if c.OutboundConfig != nil {
		outbounds = append(outbounds, *c.OutboundConfig)
	}

	if len(c.OutboundDetours) > 0 {
		outbounds = append(outbounds, c.OutboundDetours...)
	}

	if len(c.OutboundConfigs) > 0 {
		outbounds = append(outbounds, c.OutboundConfigs...)
	}

	for _, rawOutboundConfig := range outbounds {
		if c.Transport != nil {
			if rawOutboundConfig.StreamSetting == nil {
				rawOutboundConfig.StreamSetting = &StreamConfig{}
			}
			applyTransportConfig(rawOutboundConfig.StreamSetting, c.Transport)
		}
		oc, err := rawOutboundConfig.Build()
		if err != nil {
			return nil, err
		}
		config.Outbound = append(config.Outbound, oc)
	}

	return config, nil
}
