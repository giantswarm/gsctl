package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

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
	getEndpointsFunc = `
	local gsctl_out
    if gsctl_out=$(gsctl list endpoints); then
        if [[ $(echo "${gsctl_out}") != *"No endpoints configured"* ]]; then
            gsctl_out=$(echo "${gsctl_out}" | awk 'FNR > 1 {print $1}')

            COMPREPLY=( $( compgen -W "${gsctl_out}" -- "${cur}" ) )
        fi
    fi
	`

	// Default network timeout in seconds.
	timeoutSecondsDefault = 20
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{
	Use: config.ProgramName,
	// this is inherited by all child commands
	PersistentPreRunE: initConfig,
	Run:               printResult,
}

func init() {
	RootCommand.PersistentFlags().StringVarP(&flags.APIEndpoint, "endpoint", "e", "", "The API endpoint to use")

	// Use the auth token defined as an environmental variable,
	// if it exists.
	tokenFromEnv := os.Getenv("GSCTL_AUTH_TOKEN")

	RootCommand.PersistentFlags().StringVarP(&flags.Token, "auth-token", "", tokenFromEnv, "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&flags.ConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Print more information")
	RootCommand.PersistentFlags().BoolVarP(&flags.SilenceHTTPEndpointWarning, "silence-http-endpoint-warning", "", false, "Dont't print warnings when deliberately using an insecure HTTP endpoint")
	RootCommand.PersistentFlags().Int8VarP(&flags.TimeoutSeconds, "timeout", "", timeoutSecondsDefault, "Timeout for network requests, in seconds")
	RootCommand.Flags().Bool("version", false, version.Command.Short)

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

	// Custom auto-completion
	util.SetFlagBashCompletionFn(&util.BashCompletionFunc{
		Command:  RootCommand,
		Flags:    RootCommand.PersistentFlags(),
		FlagName: "endpoint",
		FnName:   "__gsctl_get_endpoints",
		FnBody:   getEndpointsFunc,
	})
	util.RegisterBashCompletionFn(RootCommand, "__gsctl_custom_func", util.GetCustomCommandCompletionFnBody())
}

// initConfig calls the config.Initialize() function
// before any command is executed (see PersistentPreRunE above).
func initConfig(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()

	var configLogger io.Writer
	if flags.SilenceHTTPEndpointWarning {
		configLogger = ioutil.Discard
	} else {
		configLogger = os.Stdout
	}

	err := config.InitializeWithLogger(fs, flags.ConfigDirPath, configLogger)
	if err != nil {
		if flags.Verbose {
			fmt.Printf("Error initializing configuration: %#v\n", err)
		}
		return microerror.Mask(err)
	}

	return nil
}

func printResult(cmd *cobra.Command, args []string) {
	isVersion, _ := cmd.Flags().GetBool("version")
	if isVersion {
		version.Command.Run(version.Command, nil)
		return
	}

	_ = cmd.Help()
}
