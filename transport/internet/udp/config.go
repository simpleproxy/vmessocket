package udp

import (
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/transport/internet"
)

const protocolName = "udp"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
