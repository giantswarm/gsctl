package main

import (
	"runtime"

	"github.com/fatih/color"
	"github.com/ryanuber/columnize"

	"github.com/giantswarm/gsctl/commands"
)

func init() {
	columnizeConfig := columnize.DefaultConfig()
	columnizeConfig.Glue = "   "

	// disable color on windows, as it is super slow
	if runtime.GOOS == "windows" {
		color.NoColor = true
	}
}

func main() {
	// top level commands
	commands.RootCommand.AddCommand(
		commands.ListCommand,
		commands.CreateCommand,
		commands.InfoCommand,
		commands.PingCommand,
		commands.LoginCommand,
		commands.LogoutCommand,
	)
	commands.RootCommand.Execute()
}
