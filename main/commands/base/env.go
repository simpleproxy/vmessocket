package base

import (
	"os"
	"path"
)

var CommandEnv CommandEnvHolder

type CommandEnvHolder struct {
	Exec string
	CommandsWidth int
}

func init() {
	exec, err := os.Executable()
	if err != nil {
		return
	}
	CommandEnv.Exec = path.Base(exec)
}
