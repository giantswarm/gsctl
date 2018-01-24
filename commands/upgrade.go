package commands

import "github.com/spf13/cobra"

var (
	// UpgradeCommand is the command to upgrade things
	UpgradeCommand = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade clusters",
		Long:  `Lets you upgrade a cluster`,
	}
)

func init() {
	RootCommand.AddCommand(UpgradeCommand)
}
