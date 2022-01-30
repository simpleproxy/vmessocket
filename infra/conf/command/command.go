package command

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/vmessocket/vmessocket/infra/conf/serial"
)

type ConfigCommand struct{}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Execute(args []string) error {
	pbConfig, err := serial.LoadJSONConfig(os.Stdin)
	if err != nil {
		return newError("failed to parse json config").Base(err)
	}

	bytesConfig, err := proto.Marshal(pbConfig)
	if err != nil {
		return newError("failed to marshal proto config").Base(err)
	}

	if _, err := os.Stdout.Write(bytesConfig); err != nil {
		return newError("failed to write proto config").Base(err)
	}
	return nil
}
