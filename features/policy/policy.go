package policy

import (
	"context"
	"runtime"
	"time"

	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/features"
)

type Timeout struct {
	Handshake      time.Duration
	ConnectionIdle time.Duration
	UplinkOnly     time.Duration
	DownlinkOnly   time.Duration
}

type Stats struct {
	UserUplink   bool
	UserDownlink bool
}

type Buffer struct {
	PerConnection int32
}

type SystemStats struct {
	InboundUplink    bool
	InboundDownlink  bool
	OutboundUplink   bool
	OutboundDownlink bool
}

type System struct {
	Stats  SystemStats
	Buffer Buffer
}

type Session struct {
	Timeouts Timeout
	Stats    Stats
	Buffer   Buffer
}

type Manager interface {
	features.Feature
	ForLevel(level uint32) Session
	ForSystem() System
}

func ManagerType() interface{} {
	return (*Manager)(nil)
}

var defaultBufferSize int32

func init() {
	const key = "vmessocket.buffer.size"
	const defaultValue = -17
	size := platform.EnvFlag{
		Name:    key,
		AltName: platform.NormalizeEnvName(key),
	}.GetValueAsInt(defaultValue)

	switch size {
	case 0:
		defaultBufferSize = -1
	case defaultValue:
		switch runtime.GOARCH {
		case "arm", "mips", "mipsle":
			defaultBufferSize = 0
		case "arm64", "mips64", "mips64le":
			defaultBufferSize = 4 * 1024
		default:
			defaultBufferSize = 512 * 1024
		}
	default:
		defaultBufferSize = int32(size) * 1024 * 1024
	}
}

func defaultBufferPolicy() Buffer {
	return Buffer{
		PerConnection: defaultBufferSize,
	}
}

func SessionDefault() Session {
	return Session{
		Timeouts: Timeout{
			Handshake:      time.Second * 60,
			ConnectionIdle: time.Second * 300,
			UplinkOnly:     time.Second * 1,
			DownlinkOnly:   time.Second * 1,
		},
		Stats: Stats{
			UserUplink:   false,
			UserDownlink: false,
		},
		Buffer: defaultBufferPolicy(),
	}
}

type policyKey int32

const (
	bufferPolicyKey policyKey = 0
)

func ContextWithBufferPolicy(ctx context.Context, p Buffer) context.Context {
	return context.WithValue(ctx, bufferPolicyKey, p)
}

func BufferPolicyFromContext(ctx context.Context) Buffer {
	pPolicy := ctx.Value(bufferPolicyKey)
	if pPolicy == nil {
		return defaultBufferPolicy()
	}
	return pPolicy.(Buffer)
}
