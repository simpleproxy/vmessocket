package conf

import (
	"net"
	"strings"

	"github.com/golang/protobuf/proto"

	v2net "github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/protocol"
	"github.com/vmessocket/vmessocket/proxy/freedom"
)

type FreedomConfig struct {
	DomainStrategy string  `json:"domainStrategy"`
	Timeout        *uint32 `json:"timeout"`
	Redirect       string  `json:"redirect"`
	UserLevel      uint32  `json:"userLevel"`
}

func (c *FreedomConfig) Build() (proto.Message, error) {
	config := new(freedom.Config)
	config.DomainStrategy = freedom.Config_AS_IS
	switch strings.ToLower(c.DomainStrategy) {
	case "useip", "use_ip", "use-ip":
		config.DomainStrategy = freedom.Config_USE_IP
	}
	if c.Timeout != nil {
		config.Timeout = *c.Timeout
	}
	config.UserLevel = c.UserLevel
	if len(c.Redirect) > 0 {
		host, portStr, err := net.SplitHostPort(c.Redirect)
		if err != nil {
			return nil, newError("invalid redirect address: ", c.Redirect, ": ", err).Base(err)
		}
		port, err := v2net.PortFromString(portStr)
		if err != nil {
			return nil, newError("invalid redirect port: ", c.Redirect, ": ", err).Base(err)
		}
		config.DestinationOverride = &freedom.DestinationOverride{
			Server: &protocol.ServerEndpoint{
				Port: uint32(port),
			},
		}
		if len(host) > 0 {
			config.DestinationOverride.Server.Address = v2net.NewIPOrDomain(v2net.ParseAddress(host))
		}
	}
	return config, nil
}
