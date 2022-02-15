// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package base defines shared basic pieces of the commands,
// in particular logging and the Command structure.
package base

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Command struct {
	Run func(cmd *Command, args []string)
	UsageLine string
	Short string
	Long string
	Flag flag.FlagSet
	CustomFlags bool
	Commands []*Command
}

func (c *Command) LongName() string {
	name := c.UsageLine
	if i := strings.Index(name, " ["); i >= 0 {
		name = strings.TrimSpace(name[:i])
	}
	if i := strings.Index(name, " "); i >= 0 {
		name = name[i+1:]
	} else {
		name = ""
	}
	return strings.TrimSpace(name)
}

func (c *Command) Name() string {
	name := c.LongName()
	if i := strings.LastIndex(name, " "); i >= 0 {
		name = name[i+1:]
	}
	return strings.TrimSpace(name)
}

func (c *Command) Usage() {
	buildCommandText(c)
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "Run '%s help %s' for details.\n", CommandEnv.Exec, c.LongName())
	SetExitStatus(2)
	Exit()
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}

func Exit() {
	os.Exit(exitStatus)
}

func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	Exit()
}

func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	SetExitStatus(1)
}

func ExitIfErrors() {
	if exitStatus != 0 {
		Exit()
	}
}

var (
	exitStatus = 0
	exitMu     sync.Mutex
)

func SetExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func GetExitStatus() int {
	return exitStatus
}
