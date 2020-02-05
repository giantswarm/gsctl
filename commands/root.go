package commands

import (
	"fmt"

	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/create"
	deletecmd "github.com/giantswarm/gsctl/commands/delete"
	"github.com/giantswarm/gsctl/commands/info"
	"github.com/giantswarm/gsctl/commands/list"
	"github.com/giantswarm/gsctl/commands/login"
	"github.com/giantswarm/gsctl/commands/logout"
	"github.com/giantswarm/gsctl/commands/ping"
	"github.com/giantswarm/gsctl/commands/scale"
	selectcmd "github.com/giantswarm/gsctl/commands/select"
	"github.com/giantswarm/gsctl/commands/show"
	"github.com/giantswarm/gsctl/commands/update"
	"github.com/giantswarm/gsctl/commands/upgrade"
	"github.com/giantswarm/gsctl/commands/version"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/util"
)

const (
	default_bash_completion_func = `
		if [[ ${last_command} -eq "gsctl_select_endpoint" ]]; then
			__gsctl_get_endpoints;
		fi
	`

	get_endpoints_func = `
	local gsctl_out
	if gsctl_out=$(gsctl list endpoints | awk 'FNR > 1 {print $1}'); then
					COMPREPLY=( $( compgen -W "${gsctl_out}" -- "${cur}" ) )
	fi
	`
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{
	Use: config.ProgramName,
	// this is inherited by all child commands
	PersistentPreRunE: initConfig,
}

func init() {
	RootCommand.PersistentFlags().StringVarP(&flags.APIEndpoint, "endpoint", "e", "", "The API endpoint to use")
	util.SetBashCompletionFunction(util.BashCompletionFunc{
		Command:  RootCommand,
		Flags:    RootCommand.PersistentFlags(),
		FlagName: "endpoint",
		FnName:   "__gsctl_get_endpoints",
		FnBody:   get_endpoints_func,
	})

	RootCommand.PersistentFlags().StringVarP(&flags.Token, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&flags.ConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Print more information")

	// add subcommands
	RootCommand.AddCommand(CompletionCommand)
	RootCommand.AddCommand(create.Command)
	RootCommand.AddCommand(deletecmd.Command)
	RootCommand.AddCommand(info.Command)
	RootCommand.AddCommand(list.Command)
	RootCommand.AddCommand(login.Command)
	RootCommand.AddCommand(logout.Command)
	RootCommand.AddCommand(ping.Command)
	RootCommand.AddCommand(scale.Command)
	RootCommand.AddCommand(selectcmd.Command)
	RootCommand.AddCommand(show.Command)
	RootCommand.AddCommand(update.Command)
	RootCommand.AddCommand(upgrade.Command)
	RootCommand.AddCommand(version.Command)

	util.RegisterBashCompletionFunction(RootCommand, "__gsctl_custom_func", default_bash_completion_func)
}

// initConfig calls the config.Initialize() function
// before any command is executed (see PersistentPreRunE above).
func initConfig(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	err := config.Initialize(fs, flags.ConfigDirPath)
	if err != nil {
		if flags.Verbose {
			fmt.Printf("Error initializing configuration: %#v\n", err)
		}
		return microerror.Mask(err)
	}

	return nil
}
