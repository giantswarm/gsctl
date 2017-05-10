package commands

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

// RootCommand is the main command of the CLI
var RootCommand = &cobra.Command{Use: config.ProgramName}

func init() {
	RootCommand.PersistentFlags().StringVarP(&cmdAPIEndpoint, "api-endpoint", "", config.DefaultAPIEndpoint, "The URL base path, to access the API")
	RootCommand.PersistentFlags().StringVarP(&cmdToken, "auth-token", "", "", "Authorization token to use for one command execution")
	RootCommand.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")
}
