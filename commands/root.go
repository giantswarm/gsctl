package commands

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{
	Use: config.ProgramName,
	// this is inherited by all child commands
	PersistentPreRunE: initConfig,
}

func init() {
	// Replaced by "endpoint" flag
	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "api-endpoint", "", "", "The URL base path, to access the API (deprecated)")
	RootCommand.PersistentFlags().MarkDeprecated("api-endpoint", "please use --endpoint or -e instead.")

	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "endpoint", "e", "", "The API endpoint to use")
	RootCommand.PersistentFlags().StringVarP(&cmdToken, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&cmdScheme, "auth-scheme", "", "giantswarm", "Authorization scheme to use (giantswarm or Bearer, case sensitive)")
	RootCommand.PersistentFlags().StringVarP(&cmdConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")
}

// initConfig calls the config.Initialize() function
// before any command is executed.
func initConfig(cmd *cobra.Command, args []string) error {
	return config.Initialize(cmdConfigDirPath)
}
