package delete

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/delete/cluster"
	"github.com/giantswarm/gsctl/commands/delete/endpoint"
	"github.com/giantswarm/gsctl/commands/delete/nodepool"
)

var (
	// Command is the command to remove things
	Command = &cobra.Command{
		Use:   "delete",
		Short: "Delete things",
		Long:  `Lets you delete a cluster, a node pool, or an API endpoint`,
	}
)

func init() {
	Command.AddCommand(cluster.Command)
	Command.AddCommand(nodepool.Command)
	Command.AddCommand(endpoint.Command)
}
