package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/vmessocket/vmessocket/common/cmdarg"
	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/core"
	_ "github.com/vmessocket/vmessocket/main/all"
	"github.com/vmessocket/vmessocket/main/commands"
)

func main() {
	flag.Parse()
	printVersion()
	if *version {
		return
	}
	server, err := startVmessocket()
	if err != nil {
		fmt.Println(err)
		os.Exit(23)
	}
	if *test {
		fmt.Println("Configuration OK.")
		os.Exit(0)
	}
	if err := server.Start(); err != nil {
		fmt.Println("Failed to start", err)
		os.Exit(-1)
	}
	defer server.Close()
	runtime.GC()
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
	}
}
