package core

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"runtime"

	"github.com/vmessocket/vmessocket/common/serial"
)

var (
	version  = "1.0.2"
	build    = "Custom"
	codename = "VMESSOCKET, an implementation of vmess and websocket protocol."
)

func Version() string {
	return version
}

func VersionStatement() []string {
	return []string{
		serial.Concat("VMESSOCKET ", Version(), " (", codename, ") ", build, " (", runtime.Version(), " ", runtime.GOOS, "/", runtime.GOARCH, ")"),
	}
}
