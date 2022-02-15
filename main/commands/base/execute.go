package base

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func Execute() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		PrintUsage(os.Stderr, RootCommand)
		return
	}
	cmdName := args[0]
	if args[0] == "help" {
		Help(os.Stdout, args[1:])
		return
	}

BigCmdLoop:
	for bigCmd := RootCommand; ; {
		for _, cmd := range bigCmd.Commands {
			if cmd.Name() != args[0] {
				continue
			}
			if len(cmd.Commands) > 0 {
				bigCmd = cmd
				args = args[1:]
				if len(args) == 0 {
					PrintUsage(os.Stderr, bigCmd)
					SetExitStatus(2)
					Exit()
				}
				if args[0] == "help" {
					Help(os.Stdout, append(strings.Split(cmdName, " "), args[1:]...))
					return
				}
				cmdName += " " + args[0]
				continue BigCmdLoop
			}
			if !cmd.Runnable() {
				continue
			}
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}

			buildCommandText(cmd)
			cmd.Run(cmd, args)
			Exit()
			return
		}
		helpArg := ""
		if i := strings.LastIndex(cmdName, " "); i >= 0 {
			helpArg = " " + cmdName[:i]
		}
		fmt.Fprintf(os.Stderr, "%s %s: unknown command\nRun '%s help%s' for usage.\n", CommandEnv.Exec, cmdName, CommandEnv.Exec, helpArg)
		SetExitStatus(2)
		Exit()
	}
}

func SortCommands() {
	sort.Slice(RootCommand.Commands, func(i, j int) bool {
		return SortLessFunc(RootCommand.Commands[i], RootCommand.Commands[j])
	})
}

var SortLessFunc = func(i, j *Command) bool {
	return i.Name() < j.Name()
}
