package http

import (
	"bytes"
	"errors"
	"strings"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
)

const (
	HTTP1 version = iota
	HTTP2
)

var (
	methods = [...]string{"get", "post", "head", "put", "delete", "options", "connect"}
	errNotHTTPMethod = errors.New("not an HTTP method")
)

type SniffHeader struct {
	version version
	host    string
}

type version byte

func beginWithHTTPMethod(b []byte) error {
	for _, m := range &methods {
		if len(b) >= len(m) && strings.EqualFold(string(b[:len(m)]), m) {
			return nil
		}

		if len(b) < len(m) {
			return common.ErrNoClue
		}
	}
	return errNotHTTPMethod
}

func SniffHTTP(b []byte) (*SniffHeader, error) {
	if err := beginWithHTTPMethod(b); err != nil {
		return nil, err
	}
	sh := &SniffHeader{
		version: HTTP1,
	}
	headers := bytes.Split(b, []byte{'\n'})
	for i := 1; i < len(headers); i++ {
		header := headers[i]
		if len(header) == 0 {
			break
		}
		parts := bytes.SplitN(header, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(string(parts[0]))
		if key == "host" {
			rawHost := strings.ToLower(string(bytes.TrimSpace(parts[1])))
			dest, err := ParseHost(rawHost, net.Port(80))
			if err != nil {
				return nil, err
			}
			sh.host = dest.Address.String()
		}
	}
	if len(sh.host) > 0 {
		return sh, nil
	}
	return nil, common.ErrNoClue
}

func (h *SniffHeader) Domain() string {
	return h.host
}

func (h *SniffHeader) Protocol() string {
	switch h.version {
	case HTTP1:
		return "http1"
	case HTTP2:
		return "http2"
	default:
		return "unknown"
	}
}
