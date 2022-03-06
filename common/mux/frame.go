package mux

import (
	"github.com/vmessocket/vmessocket/common/bitmask"
)

const (
	OptionData             bitmask.Byte  = 0x01
	OptionError            bitmask.Byte  = 0x02
	SessionStatusNew       SessionStatus = 0x01
	SessionStatusKeep      SessionStatus = 0x02
	SessionStatusEnd       SessionStatus = 0x03
	SessionStatusKeepAlive SessionStatus = 0x04
)

type SessionStatus byte
