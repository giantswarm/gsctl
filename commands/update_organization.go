package commands

import "github.com/spf13/cobra"

var (

	// UpdateOrganizationCommand performs the "update organization" function
	UpdateOrganizationCommand = &cobra.Command{
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
	UpdateOrganizationCommand.Flags().StringVarP(&cmdOrganization, "organization", "o", "", "ID of the organization to modify")
	UpdateCommand.AddCommand(UpdateOrganizationCommand)
}
