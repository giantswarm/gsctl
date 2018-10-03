package commands

import "github.com/spf13/cobra"

var (
	// ShowCommand is the command to display single items
	ShowCommand = &cobra.Command{
		Use:   "show",
		Short: "Show things, like clusters, releases",
		Long:  `Print details of a cluster or a release`,
	}
)

func init() {
	RootCommand.AddCommand(ShowCommand)
}
