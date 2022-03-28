package websocket

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/transport/internet"
)

type dialerWithEarlyData struct {
	dialer  *websocket.Dialer
	uriBase string
	config  *Config
}

type dialerWithEarlyDataRelayed struct {
	uriBase string
	config  *Config
}

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))
	conn, err := dialWebsocket(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial WebSocket").Base(err)
	}
	return internet.Connection(conn), nil
}

func dialWebsocket(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	wsSettings := streamSettings.ProtocolSettings.(*Config)
	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return internet.DialSystem(ctx, dest, streamSettings.SocketSettings)
		},
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		HandshakeTimeout: time.Second * 8,
	}
	protocol := "ws"
	host := dest.NetAddr()
	if (protocol == "ws" && dest.Port == 80) || (protocol == "wss" && dest.Port == 443) {
		host = dest.Address.String()
	}
	uri := protocol + "://" + host + wsSettings.GetNormalizedPath()
	conn, resp, err := dialer.Dial(uri, wsSettings.GetRequestHeader())
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", uri, "): ", reason).Base(err)
	}
	return newConnection(conn, conn.RemoteAddr()), nil
}

func (d dialerWithEarlyData) Dial(earlyData []byte) (*websocket.Conn, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)
	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}
	dialFunction := func() (*websocket.Conn, *http.Response, error) {
		return d.dialer.Dial(d.uriBase+earlyDataBuf.String(), d.config.GetRequestHeader())
	}
	conn, resp, err := dialFunction()
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", d.uriBase, ") with early data: ", reason).Base(err)
	}
	return conn, nil
}

func (d dialerWithEarlyDataRelayed) Dial(earlyData []byte) (io.ReadWriteCloser, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)
	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}
	return nil, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
