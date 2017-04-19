package commands

import "github.com/spf13/cobra"

var (
	// ListCommand is the command to list things
	ListCommand = &cobra.Command{
		Use:   "list",
		Short: "List things, like organizations, clusters, key-pairs",
		Long:  `Prints a list of the things you have access to`,
	}
)

func init() {
	RootCommand.AddCommand(ListCommand)
}
