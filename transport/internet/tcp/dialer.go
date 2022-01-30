//go:build !confonly
// +build !confonly

package tcp

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/transport/internet"
	"github.com/vmessocket/vmessocket/transport/internet/tls"
)

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("dialing TCP to ", dest).WriteToLog(session.ExportIDToError(ctx))
	conn, err := internet.DialSystem(ctx, dest, streamSettings.SocketSettings)
	if err != nil {
		return nil, err
	}

	if config := tls.ConfigFromStreamSettings(streamSettings); config != nil {
		tlsConfig := config.GetTLSConfig(tls.WithDestination(dest))
		conn = tls.Client(conn, tlsConfig)
	}

	tcpSettings := streamSettings.ProtocolSettings.(*Config)
	if tcpSettings.HeaderSettings != nil {
		headerConfig, err := tcpSettings.HeaderSettings.GetInstance()
		if err != nil {
			return nil, newError("failed to get header settings").Base(err).AtError()
		}
		auth, err := internet.CreateConnectionAuthenticator(headerConfig)
		if err != nil {
			return nil, newError("failed to create header authenticator").Base(err).AtError()
		}
		conn = auth.Client(conn)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
