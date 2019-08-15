package update

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/update/nodepool"
	"github.com/giantswarm/gsctl/commands/update/organization"
)

var (
	// Command is the command to modify resources
	Command = &cobra.Command{
		Use:   "update",
		Short: "Modify node pool or organization details",
		Long:  `Modify details of a node pool or an organization`,
	}
)

func init() {
	Command.AddCommand(organization.Command)
	Command.AddCommand(nodepool.Command)
}
