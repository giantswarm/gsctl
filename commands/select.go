package commands

import "github.com/spf13/cobra"

var (
	// SelectCommand is the command to list things
	SelectCommand = &cobra.Command{
		Use:   "select",
		Short: "Select things, like the API endpoint to use",
		Long:  `Select things, like the API endpoint to use`,
	}
)

func init() {
	RootCommand.AddCommand(SelectCommand)
}
