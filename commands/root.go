package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
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
	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "api-endpoint", "", "", "The URL base path, to access the API (deprecated)")
	RootCommand.PersistentFlags().MarkDeprecated("api-endpoint", "please use --endpoint or -e instead.")

	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "endpoint", "e", "", "The API endpoint to use")
	RootCommand.PersistentFlags().StringVarP(&cmdToken, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&cmdConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")
}

// initConfig calls the config.Initialize() function
// before any command is executed.
func initConfig(cmd *cobra.Command, args []string) error {
	err := config.Initialize(cmdConfigDirPath)
	if err != nil {
		return microerror.Mask(err)
	}

	err = initClient()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func initClient() error {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	ClientConfig = &client.Configuration{
		AuthHeader: scheme + " " + token,
		Endpoint:   endpoint,
		Timeout:    10 * time.Second,
		UserAgent:  config.UserAgent(),
	}

	var err error
	ClientV2, err = client.NewV2(ClientConfig)
	if err != nil {
		return microerror.Maskf(couldNotCreateClientError, err.Error())
	}

	return nil
}
