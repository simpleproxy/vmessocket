package conf

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/vmessocket/vmessocket/common/platform/filesystem"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/transport/internet"
	"github.com/vmessocket/vmessocket/transport/internet/tcp"
	"github.com/vmessocket/vmessocket/transport/internet/websocket"
)

type ProxyConfig struct {
	TransportLayerProxy bool   `json:"transportLayer"`
}

type StreamConfig struct {
	Network        *TransportProtocol `json:"network"`
	Security       string             `json:"security"`
	TCPSettings    *TCPConfig         `json:"tcpSettings"`
	WSSettings     *WebSocketConfig   `json:"wsSettings"`
}

type TCPConfig struct {
	HeaderConfig json.RawMessage `json:"header"`
}

type TLSCertConfig struct {
	CertFile string   `json:"certificateFile"`
	CertStr  []string `json:"certificate"`
	KeyFile  string   `json:"keyFile"`
	KeyStr   []string `json:"key"`
	Usage    string   `json:"usage"`
}

type TransportProtocol string

type WebSocketConfig struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
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
	return config, nil
}

func (c *TCPConfig) Build() (proto.Message, error) {
	config := new(tcp.Config)
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
		Path:   path,
		Header: header,
	}
	return config, nil
}
