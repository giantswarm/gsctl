package organization

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/update/organization/setcredentials"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command performs the "update organization" function
	Command = &cobra.Command{
		Use:     "organization",
		Aliases: []string{"org"},
		Short:   "Modify organization details",
		Long: `Modify details of an organization

Examples:

  gsctl update organization set-credentials -o acme ...
`,
	}
)

func init() {
	Command.Flags().StringVarP(&flags.CmdOrganizationID, "organization", "o", "", "ID of the organization to modify")

	Command.AddCommand(setcredentials.Command)
}
