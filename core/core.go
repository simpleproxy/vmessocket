package core

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"runtime"

	"github.com/vmessocket/vmessocket/common/serial"
)

var (
	version  = "1.0.0"
	build    = "Custom"
	codename = "VMESSOCKET, an implementation of vmess and websocket protocol."
	intro    = "A unified platform for anti-censorship."
)

func Version() string {
	return version
}

func VersionStatement() []string {
	return []string{
		serial.Concat("VMESSOCKET ", Version(), " (", codename, ") ", build, " (", runtime.Version(), " ", runtime.GOOS, "/", runtime.GOARCH, ")"),
		intro,
	}
}
