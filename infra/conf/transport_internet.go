package conf

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/vmessocket/vmessocket/common/platform/filesystem"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/infra/conf/cfgcommon"
	"github.com/vmessocket/vmessocket/transport/internet"
	"github.com/vmessocket/vmessocket/transport/internet/http"
	"github.com/vmessocket/vmessocket/transport/internet/tcp"
	"github.com/vmessocket/vmessocket/transport/internet/tls"
	"github.com/vmessocket/vmessocket/transport/internet/websocket"
)

type HTTPConfig struct {
	Host    *cfgcommon.StringList            `json:"host"`
	Path    string                           `json:"path"`
	Method  string                           `json:"method"`
	Headers map[string]*cfgcommon.StringList `json:"headers"`
}

type ProxyConfig struct {
	Tag                 string `json:"tag"`
	TransportLayerProxy bool   `json:"transportLayer"`
}

type SocketConfig struct {
	Mark                 uint32 `json:"mark"`
	TFO                  *bool  `json:"tcpFastOpen"`
	TFOQueueLength       uint32 `json:"tcpFastOpenQueueLength"`
	TProxy               string `json:"tproxy"`
	AcceptProxyProtocol  bool   `json:"acceptProxyProtocol"`
	TCPKeepAliveInterval int32  `json:"tcpKeepAliveInterval"`
}

type StreamConfig struct {
	Network        *TransportProtocol  `json:"network"`
	Security       string              `json:"security"`
	TLSSettings    *TLSConfig          `json:"tlsSettings"`
	TCPSettings    *TCPConfig          `json:"tcpSettings"`
	WSSettings     *WebSocketConfig    `json:"wsSettings"`
	HTTPSettings   *HTTPConfig         `json:"httpSettings"`
	SocketSettings *SocketConfig       `json:"sockopt"`
}

type TCPConfig struct {
	HeaderConfig        json.RawMessage `json:"header"`
	AcceptProxyProtocol bool            `json:"acceptProxyProtocol"`
}

type TLSCertConfig struct {
	CertFile string   `json:"certificateFile"`
	CertStr  []string `json:"certificate"`
	KeyFile  string   `json:"keyFile"`
	KeyStr   []string `json:"key"`
	Usage    string   `json:"usage"`
}

type TLSConfig struct {
	Insecure                         bool                  `json:"allowInsecure"`
	Certs                            []*TLSCertConfig      `json:"certificates"`
	ServerName                       string                `json:"serverName"`
	ALPN                             *cfgcommon.StringList `json:"alpn"`
	EnableSessionResumption          bool                  `json:"enableSessionResumption"`
	DisableSystemRoot                bool                  `json:"disableSystemRoot"`
	PinnedPeerCertificateChainSha256 *[]string             `json:"pinnedPeerCertificateChainSha256"`
	VerifyClientCertificate          bool                  `json:"verifyClientCertificate"`
}

type TransportProtocol string

type WebSocketConfig struct {
	Path                 string            `json:"path"`
	Headers              map[string]string `json:"headers"`
	AcceptProxyProtocol  bool              `json:"acceptProxyProtocol"`
	MaxEarlyData         int32             `json:"maxEarlyData"`
	UseBrowserForwarding bool              `json:"useBrowserForwarding"`
	EarlyDataHeaderName  string            `json:"earlyDataHeaderName"`
}

func readFileOrString(f string, s []string) ([]byte, error) {
	if len(f) > 0 {
		return filesystem.ReadFile(f)
	}
	if len(s) > 0 {
		return []byte(strings.Join(s, "\n")), nil
	}
	return nil, newError("both file and bytes are empty.")
}

func (c *HTTPConfig) Build() (proto.Message, error) {
	config := &http.Config{
		Path: c.Path,
	}
	if c.Host != nil {
		config.Host = []string(*c.Host)
	}
	if c.Method != "" {
		config.Method = c.Method
	}
	return config, nil
}

func (v *ProxyConfig) Build() (*internet.ProxyConfig, error) {
	if v.Tag == "" {
		return nil, newError("Proxy tag is not set.")
	}
	return &internet.ProxyConfig{
		Tag:                 v.Tag,
		TransportLayerProxy: v.TransportLayerProxy,
	}, nil
}

func (c *SocketConfig) Build() (*internet.SocketConfig, error) {
	var tfoSettings internet.SocketConfig_TCPFastOpenState
	if c.TFO != nil {
		if *c.TFO {
			tfoSettings = internet.SocketConfig_Enable
		} else {
			tfoSettings = internet.SocketConfig_Disable
		}
	}
	tfoQueueLength := c.TFOQueueLength
	if tfoQueueLength == 0 {
		tfoQueueLength = 4096
	}
	var tproxy internet.SocketConfig_TProxyMode
	switch strings.ToLower(c.TProxy) {
	case "tproxy":
		tproxy = internet.SocketConfig_TProxy
	case "redirect":
		tproxy = internet.SocketConfig_Redirect
	default:
		tproxy = internet.SocketConfig_Off
	}
	return &internet.SocketConfig{
		Mark:                 c.Mark,
		Tfo:                  tfoSettings,
		TfoQueueLength:       tfoQueueLength,
		Tproxy:               tproxy,
		AcceptProxyProtocol:  c.AcceptProxyProtocol,
		TcpKeepAliveInterval: c.TCPKeepAliveInterval,
	}, nil
}

func (c *StreamConfig) Build() (*internet.StreamConfig, error) {
	config := &internet.StreamConfig{
		ProtocolName: "tcp",
	}
	if c.Network != nil {
		protocol, err := c.Network.Build()
		if err != nil {
			return nil, err
		}
		config.ProtocolName = protocol
	}
	if strings.EqualFold(c.Security, "tls") {
		tlsSettings := c.TLSSettings
		if tlsSettings == nil {
			tlsSettings = &TLSConfig{}
		}
		ts, err := tlsSettings.Build()
		if err != nil {
			return nil, newError("Failed to build TLS config.").Base(err)
		}
		tm := serial.ToTypedMessage(ts)
		config.SecuritySettings = append(config.SecuritySettings, tm)
		config.SecurityType = tm.Type
	}
	if c.TCPSettings != nil {
		ts, err := c.TCPSettings.Build()
		if err != nil {
			return nil, newError("Failed to build TCP config.").Base(err)
		}
		config.TransportSettings = append(config.TransportSettings, &internet.TransportConfig{
			ProtocolName: "tcp",
			Settings:     serial.ToTypedMessage(ts),
		})
	}
	if c.WSSettings != nil {
		ts, err := c.WSSettings.Build()
		if err != nil {
			return nil, newError("Failed to build WebSocket config.").Base(err)
		}
		config.TransportSettings = append(config.TransportSettings, &internet.TransportConfig{
			ProtocolName: "websocket",
			Settings:     serial.ToTypedMessage(ts),
		})
	}
	if c.HTTPSettings != nil {
		ts, err := c.HTTPSettings.Build()
		if err != nil {
			return nil, newError("Failed to build HTTP config.").Base(err)
		}
		config.TransportSettings = append(config.TransportSettings, &internet.TransportConfig{
			ProtocolName: "http",
			Settings:     serial.ToTypedMessage(ts),
		})
	}
	if c.SocketSettings != nil {
		ss, err := c.SocketSettings.Build()
		if err != nil {
			return nil, newError("Failed to build sockopt.").Base(err)
		}
		config.SocketSettings = ss
	}
	return config, nil
}

func (c *TCPConfig) Build() (proto.Message, error) {
	config := new(tcp.Config)
	if c.AcceptProxyProtocol {
		config.AcceptProxyProtocol = c.AcceptProxyProtocol
	}
	return config, nil
}

func (c *TLSCertConfig) Build() (*tls.Certificate, error) {
	certificate := new(tls.Certificate)
	cert, err := readFileOrString(c.CertFile, c.CertStr)
	if err != nil {
		return nil, newError("failed to parse certificate").Base(err)
	}
	certificate.Certificate = cert
	if len(c.KeyFile) > 0 || len(c.KeyStr) > 0 {
		key, err := readFileOrString(c.KeyFile, c.KeyStr)
		if err != nil {
			return nil, newError("failed to parse key").Base(err)
		}
		certificate.Key = key
	}
	switch strings.ToLower(c.Usage) {
	case "encipherment":
		certificate.Usage = tls.Certificate_ENCIPHERMENT
	case "verify":
		certificate.Usage = tls.Certificate_AUTHORITY_VERIFY
	case "verifyclient":
		certificate.Usage = tls.Certificate_AUTHORITY_VERIFY_CLIENT
	case "issue":
		certificate.Usage = tls.Certificate_AUTHORITY_ISSUE
	default:
		certificate.Usage = tls.Certificate_ENCIPHERMENT
	}
	return certificate, nil
}

func (c *TLSConfig) Build() (proto.Message, error) {
	config := new(tls.Config)
	config.Certificate = make([]*tls.Certificate, len(c.Certs))
	for idx, certConf := range c.Certs {
		cert, err := certConf.Build()
		if err != nil {
			return nil, err
		}
		config.Certificate[idx] = cert
	}
	serverName := c.ServerName
	config.AllowInsecure = c.Insecure
	config.VerifyClientCertificate = c.VerifyClientCertificate
	if len(c.ServerName) > 0 {
		config.ServerName = serverName
	}
	if c.ALPN != nil && len(*c.ALPN) > 0 {
		config.NextProtocol = []string(*c.ALPN)
	}
	config.EnableSessionResumption = c.EnableSessionResumption
	config.DisableSystemRoot = c.DisableSystemRoot
	if c.PinnedPeerCertificateChainSha256 != nil {
		config.PinnedPeerCertificateChainSha256 = [][]byte{}
		for _, v := range *c.PinnedPeerCertificateChainSha256 {
			hashValue, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return nil, err
			}
			config.PinnedPeerCertificateChainSha256 = append(config.PinnedPeerCertificateChainSha256, hashValue)
		}
	}
	return config, nil
}

func (p TransportProtocol) Build() (string, error) {
	switch strings.ToLower(string(p)) {
	case "tcp":
		return "tcp", nil
	case "ws", "websocket":
		return "websocket", nil
	case "h2", "http":
		return "http", nil
	default:
		return "", newError("Config: unknown transport protocol: ", p)
	}
}

func (c *WebSocketConfig) Build() (proto.Message, error) {
	path := c.Path
	header := make([]*websocket.Header, 0, 32)
	for key, value := range c.Headers {
		header = append(header, &websocket.Header{
			Key:   key,
			Value: value,
		})
	}
	config := &websocket.Config{
		Path:                 path,
		Header:               header,
		MaxEarlyData:         c.MaxEarlyData,
		UseBrowserForwarding: c.UseBrowserForwarding,
		EarlyDataHeaderName:  c.EarlyDataHeaderName,
	}
	if c.AcceptProxyProtocol {
		config.AcceptProxyProtocol = c.AcceptProxyProtocol
	}
	return config, nil
}
