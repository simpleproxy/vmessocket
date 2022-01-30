package transport

import (
	"github.com/vmessocket/vmessocket/transport/internet"
)

func (c *Config) Apply() error {
	if c == nil {
		return nil
	}
	return internet.ApplyGlobalTransportSettings(c.TransportSettings)
}
