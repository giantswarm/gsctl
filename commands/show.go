package commands

import "github.com/spf13/cobra"

var (
	// ShowCommand is the command to display single items
	ShowCommand = &cobra.Command{
		Use:   "show",
		Short: "Access cluster details",
		Long:  `Print details of a cluster`,
	}
)

func init() {
	RootCommand.AddCommand(ShowCommand)
}
