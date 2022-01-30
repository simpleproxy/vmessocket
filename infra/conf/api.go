package conf

import (
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"

	"github.com/vmessocket/vmessocket/app/commander"
	loggerservice "github.com/vmessocket/vmessocket/app/log/command"
	handlerservice "github.com/vmessocket/vmessocket/app/proxyman/command"
	"github.com/vmessocket/vmessocket/common/serial"
)

type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

func (c *APIConfig) Build() (*commander.Config, error) {
	if c.Tag == "" {
		return nil, newError("API tag can't be empty.")
	}

	services := make([]*serial.TypedMessage, 0, 16)
	for _, s := range c.Services {
		switch strings.ToLower(s) {
		case "reflectionservice":
			services = append(services, serial.ToTypedMessage(&commander.ReflectionConfig{}))
		case "handlerservice":
			services = append(services, serial.ToTypedMessage(&handlerservice.Config{}))
		case "loggerservice":
			services = append(services, serial.ToTypedMessage(&loggerservice.Config{}))
		default:
			if !strings.HasPrefix(s, "#") {
				continue
			}
			message, err := desc.LoadMessageDescriptor(s[1:])
			if err != nil || message == nil {
				return nil, newError("Cannot find API", s, "").Base(err)
			}
			serviceConfig := dynamic.NewMessage(message)
			services = append(services, serial.ToTypedMessage(serviceConfig))
		}
	}

	return &commander.Config{
		Tag:     c.Tag,
		Service: services,
	}, nil
}
