package commands

import "github.com/spf13/cobra"

var (
	// ScaleCommand is the command to scale things
	ScaleCommand = &cobra.Command{
		Use:   "scale",
		Short: "Scale clusters",
		Long:  `Lets you increase or decrease the number of worker nodes in a cluster.`,
	}
)

func init() {
	// subcommands
	RootCommand.AddCommand(ScaleCommand)
}
