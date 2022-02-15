package base

import (
	"os"
	"path"
)

type CommandEnvHolder struct {
	Exec string
	CommandsWidth int
}

var CommandEnv CommandEnvHolder

func init() {
	exec, err := os.Executable()
	if err != nil {
		return
	}
	CommandEnv.Exec = path.Base(exec)
}
