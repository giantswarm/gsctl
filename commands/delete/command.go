package delete

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/delete/cluster"
)

var (
	// Command is the command to remove things
	Command = &cobra.Command{
		Use:   "delete",
		Short: "Delete things",
		Long:  `Lets you delete a cluster`,
	}
)

func init() {
	Command.AddCommand(cluster.Command)
}
