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
	"github.com/vmessocket/vmessocket/transport/internet/tls"
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
	if config := tls.ConfigFromStreamSettings(streamSettings); config != nil {
		protocol = "wss"
		dialer.TLSClientConfig = config.GetTLSConfig(tls.WithDestination(dest), tls.WithNextProto("http/1.1"))
	}
	host := dest.NetAddr()
	if (protocol == "ws" && dest.Port == 80) || (protocol == "wss" && dest.Port == 443) {
		host = dest.Address.String()
	}
	uri := protocol + "://" + host + wsSettings.GetNormalizedPath()
	if wsSettings.MaxEarlyData != 0 {
		return newConnectionWithDelayedDial(&dialerWithEarlyData{
			dialer:  dialer,
			uriBase: uri,
			config:  wsSettings,
		}), nil
	}
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
	earlydata := bytes.NewReader(earlyData)
	limitedEarlyDatareader := io.LimitReader(earlydata, int64(d.config.MaxEarlyData))
	n, encerr := io.Copy(base64EarlyDataEncoder, limitedEarlyDatareader)
	if encerr != nil {
		return nil, newError("websocket delayed dialer cannot encode early data").Base(encerr)
	}
	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}
	dialFunction := func() (*websocket.Conn, *http.Response, error) {
		return d.dialer.Dial(d.uriBase+earlyDataBuf.String(), d.config.GetRequestHeader())
	}
	if d.config.EarlyDataHeaderName != "" {
		dialFunction = func() (*websocket.Conn, *http.Response, error) {
			earlyDataStr := earlyDataBuf.String()
			currentHeader := d.config.GetRequestHeader()
			currentHeader.Set(d.config.EarlyDataHeaderName, earlyDataStr)
			return d.dialer.Dial(d.uriBase, currentHeader)
		}
	}
	conn, resp, err := dialFunction()
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", d.uriBase, ") with early data: ", reason).Base(err)
	}
	if n != int64(len(earlyData)) {
		if errWrite := conn.WriteMessage(websocket.BinaryMessage, earlyData[n:]); errWrite != nil {
			return nil, newError("failed to dial to (", d.uriBase, ") with early data as write of remainder early data failed: ").Base(err)
		}
	}
	return conn, nil
}

func (d dialerWithEarlyDataRelayed) Dial(earlyData []byte) (io.ReadWriteCloser, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)
	earlydata := bytes.NewReader(earlyData)
	limitedEarlyDatareader := io.LimitReader(earlydata, int64(d.config.MaxEarlyData))
	_, encerr := io.Copy(base64EarlyDataEncoder, limitedEarlyDatareader)
	if encerr != nil {
		return nil, newError("websocket delayed dialer cannot encode early data").Base(encerr)
	}
	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}
	if d.config.EarlyDataHeaderName != "" {
		earlyDataStr := earlyDataBuf.String()
		currentHeader := d.config.GetRequestHeader()
		currentHeader.Set(d.config.EarlyDataHeaderName, earlyDataStr)
	}
	return nil, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
