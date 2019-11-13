package create

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/create/cluster"
	"github.com/giantswarm/gsctl/commands/create/keypair"
	"github.com/giantswarm/gsctl/commands/create/kubeconfig"
	"github.com/giantswarm/gsctl/commands/create/nodepool"
)

var (
	// Command is the command to create things.
	Command = &cobra.Command{
		Use:   "create",
		Short: "Create clusters, key pairs, node pools",
		Long:  `Lets you create things like clusters, key pairs or kubectl configuration files`,
	}
)

func init() {
	Command.AddCommand(cluster.Command)
	Command.AddCommand(keypair.Command)
	Command.AddCommand(kubeconfig.Command)
	Command.AddCommand(nodepool.Command)
}
