package commands

import "github.com/spf13/cobra"

var (
	// UpdateCommand is the command to modify resources
	UpdateCommand = &cobra.Command{
		Use:   "update",
		Short: "Modify organization details",
		Long:  `Modify details of an organization`,
	}
)

func init() {
	RootCommand.AddCommand(UpdateCommand)
}
