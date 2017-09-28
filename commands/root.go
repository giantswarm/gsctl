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
	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "api-endpoint", "", "", "The URL base path, to access the API")
	RootCommand.PersistentFlags().StringVarP(&cmdToken, "auth-token", "", "", "Authorization token to use")
	RootCommand.PersistentFlags().StringVarP(&cmdConfigDirPath, "config-dir", "", config.DefaultConfigDirPath, "Configuration directory path to use")
	RootCommand.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")
}

// initConfig calls the config.Initialize() function
// before any command is executed.
func initConfig(cmd *cobra.Command, args []string) error {
	return config.Initialize(cmdConfigDirPath)
}
