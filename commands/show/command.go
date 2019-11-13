package show

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/show/cluster"
	"github.com/giantswarm/gsctl/commands/show/nodepool"
	"github.com/giantswarm/gsctl/commands/show/release"
)

var (
	// Command is the command to display single items
	Command = &cobra.Command{
		Use:   "show",
		Short: "Show clusters, node pools, releases",
		Long:  `Print details of a cluster or a release`,
	}
)

func init() {
	Command.AddCommand(cluster.ShowClusterCommand)
	Command.AddCommand(nodepool.ShowNodepoolCommand)
	Command.AddCommand(release.ShowReleaseCommand)
}
