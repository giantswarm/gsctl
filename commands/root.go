package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/create"
	deletecmd "github.com/giantswarm/gsctl/commands/delete"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/list"
	"github.com/giantswarm/gsctl/commands/show"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/microerror"
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{
	Use: config.ProgramName,
	// this is inherited by all child commands
	PersistentPreRunE: initConfig,
}

var (
	// ClientV2 is the latest client wrapper we create for all commands to use
	ClientV2 *client.WrapperV2

	// ClientConfig is a client configuration we apply when creating a new clients
	ClientConfig *client.Configuration
)

func init() {
	// Replaced by "endpoint" flag
	RootCommand.PersistentFlags().StringVarP(&flags.CmdAPIEndpoint, "api-endpoint", "", "", "The URL base path, to access the API (deprecated)")
	RootCommand.PersistentFlags().MarkDeprecated("api-endpoint", "please use --endpoint or -e instead.")

	RootCommand.PersistentFlags().StringVarP(&flags.CmdAPIEndpoint, "endpoint", "e", "", "The API endpoint to use")
	RootCommand.PersistentFlags().StringVarP(&flags.CmdToken, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&flags.CmdConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&flags.CmdVerbose, "verbose", "v", false, "Print more information")

	// add subcommands
	RootCommand.AddCommand(show.Command)
	RootCommand.AddCommand(create.Command)
	RootCommand.AddCommand(list.Command)
	RootCommand.AddCommand(deletecmd.Command)
}

// initConfig calls the config.Initialize() function
// before any command is executed.
func initConfig(cmd *cobra.Command, args []string) error {
	err := config.Initialize(flags.CmdConfigDirPath)
	if err != nil {
		return microerror.Mask(err)
	}

	err = InitClient()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// InitClient initializes the client wrapper.
// TODO: let every command initialize its own client, then remove this.
func InitClient() error {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)

	ClientConfig = &client.Configuration{
		AuthHeaderGetter: config.Config.AuthHeaderGetter(endpoint, flags.CmdToken),
		Endpoint:         endpoint,
		Timeout:          20 * time.Second,
		UserAgent:        config.UserAgent(),
	}

	var err error
	ClientV2, err = client.NewV2(ClientConfig)
	if err != nil {
		return microerror.Maskf(errors.CouldNotCreateClientError, err.Error())
	}

	return nil
}
