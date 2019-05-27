package scale

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/scale/cluster"
)

var (
	// Command is the command to scale things
	Command = &cobra.Command{
		Use:   "scale",
		Short: "Scale clusters",
		Long:  `Lets you increase or decrease the number of worker nodes in a cluster.`,
	}
)

func init() {
	// subcommands
	Command.AddCommand(cluster.Command)
}
