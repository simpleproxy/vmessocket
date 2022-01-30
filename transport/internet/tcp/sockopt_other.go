//go:build !linux && !freebsd && !confonly
// +build !linux,!freebsd,!confonly

package tcp

import (
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/transport/internet"
)

func GetOriginalDestination(conn internet.Connection) (net.Destination, error) {
	return net.Destination{}, nil
}
