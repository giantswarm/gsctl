package commands

import "github.com/spf13/cobra"

var (
	// DeleteCommand is the command to remove things
	DeleteCommand = &cobra.Command{
		Use:   "delete",
		Short: "Delete clusters",
		Long:  `Lets you delete a cluster`,
	}
)

func init() {
	RootCommand.AddCommand(DeleteCommand)
}
