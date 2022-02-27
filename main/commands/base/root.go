package base

var RootCommand *Command

func RegisterCommand(cmd *Command) {
	RootCommand.Commands = append(RootCommand.Commands, cmd)
}

func init() {
	RootCommand = &Command{
		UsageLine: CommandEnv.Exec,
		Long:      "The root command",
	}
}
