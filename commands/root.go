package commands

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/create"
	deletecmd "github.com/giantswarm/gsctl/commands/delete"
	"github.com/giantswarm/gsctl/commands/info"
	"github.com/giantswarm/gsctl/commands/list"
	"github.com/giantswarm/gsctl/commands/login"
	"github.com/giantswarm/gsctl/commands/logout"
	"github.com/giantswarm/gsctl/commands/scale"
	selectcmd "github.com/giantswarm/gsctl/commands/select"
	"github.com/giantswarm/gsctl/commands/show"
	"github.com/giantswarm/gsctl/commands/update"
	"github.com/giantswarm/gsctl/commands/upgrade"
	"github.com/giantswarm/gsctl/commands/version"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/flags"
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{
	Use: config.ProgramName,
	// this is inherited by all child commands
	PersistentPreRunE: initConfig,
}

func init() {
	RootCommand.PersistentFlags().StringVarP(&flags.CmdAPIEndpoint, "endpoint", "e", "", "The API endpoint to use")
	RootCommand.PersistentFlags().StringVarP(&flags.CmdToken, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&flags.CmdConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&flags.CmdVerbose, "verbose", "v", false, "Print more information")

	// add subcommands
	RootCommand.AddCommand(CompletionCommand)
	RootCommand.AddCommand(create.Command)
	RootCommand.AddCommand(deletecmd.Command)
	RootCommand.AddCommand(info.Command)
	RootCommand.AddCommand(list.Command)
	RootCommand.AddCommand(login.Command)
	RootCommand.AddCommand(logout.Command)
	RootCommand.AddCommand(scale.Command)
	RootCommand.AddCommand(selectcmd.Command)
	RootCommand.AddCommand(show.Command)
	RootCommand.AddCommand(update.Command)
	RootCommand.AddCommand(upgrade.Command)
	RootCommand.AddCommand(version.Command)
}

// initConfig calls the config.Initialize() function
// before any command is executed (see PersistentPreRunE above).
func initConfig(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	err := config.Initialize(fs, flags.CmdConfigDirPath)
	if err != nil {
		if flags.CmdVerbose {
			fmt.Printf("Error initializing configuration: %#v\n", err)
		}
		return microerror.Mask(err)
	}

	return nil
}
