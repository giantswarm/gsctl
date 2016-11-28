package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

var (
	// VersionCommand is the "version" go command
	VersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Print version number",
		Long: `Prints the gsctl version number.

When executed with the --verbose flag, the build date is printed in addition.`,
		Run: printVersion,
	}
)

func init() {
	RootCommand.AddCommand(VersionCommand)
}

// printInfo prints some information on the current user and configuration
func printVersion(cmd *cobra.Command, args []string) {
	if config.Version != "" {
		fmt.Println(config.Version)
	} else {
		fmt.Println("Info: version number is only available in a built binary")
	}
	if cmdVerbose {
		if config.BuildDate != "" {
			fmt.Println(config.BuildDate)
		} else {
			fmt.Println("Info: build date/time is only available in a built binary")
		}
	}
}
