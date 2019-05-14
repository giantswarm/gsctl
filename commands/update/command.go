package update

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/update/organization"
)

var (
	// Command is the command to modify resources
	Command = &cobra.Command{
		Use:   "update",
		Short: "Modify organization details",
		Long:  `Modify details of an organization`,
	}
)

func init() {
	Command.AddCommand(organization.Command)
}
