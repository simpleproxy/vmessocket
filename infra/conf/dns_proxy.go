package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/infra/conf/cfgcommon"
	"github.com/vmessocket/vmessocket/proxy/dns"
)

type DNSOutboundConfig struct {
	Network   cfgcommon.Network  `json:"network"`
	Address   *cfgcommon.Address `json:"address"`
	Port      uint16             `json:"port"`
}

func (c *DNSOutboundConfig) Build() (proto.Message, error) {
	config := &dns.Config{
		Server: &net.Endpoint{
			Network: c.Network.Build(),
			Port:    uint32(c.Port),
		},
	}
	if c.Address != nil {
		config.Server.Address = c.Address.Build()
	}
	return config, nil
}
