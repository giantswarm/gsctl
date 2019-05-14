package upgrade

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/upgrade/cluster"
)

var (
	// Command is the command to upgrade things
	Command = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade clusters",
		Long:  `Lets you upgrade a cluster`,
	}
)

func init() {
	Command.AddCommand(cluster.Command)
}
