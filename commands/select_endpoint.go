package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/microerror"

	"github.com/spf13/cobra"
)

var (
	// SelectEndpointCommand performs the "select endpoint" function
	SelectEndpointCommand = &cobra.Command{
		Use:     "endpoint <endpoint>",
		Aliases: []string{"endpoints"},
		Short:   "Select endpoint to use",
		Long: `Select the API endpoint to use in subsequent commands.

To add an endpoint for the first time, or to re-login, use the 'gsctl login'
command with that endpoint.

To find out which endpoints are selectable, use the 'gsctl list endpoints'
command.
`,
		PreRun: selectEndpointPreRun,
		Run:    selectEndpoint,
	}
)

func init() {
	SelectCommand.AddCommand(SelectEndpointCommand)
}

func selectEndpointPreRun(cmd *cobra.Command, cmdLineArgs []string) {
	err := verifySelectEndpointPreconditions(cmdLineArgs)
	if err != nil {
		headline := ""
		subtext := ""

		switch {
		case err.Error() == "":
			headline = "Unknown error."
		case IsEndpointMissingError(err):
			headline = "No endpoint specified."
			subtext = "Please give an endpoint URL to use. Use --help for details."
		default:
			headline = err.Error()
		}

		// print output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}
}

func verifySelectEndpointPreconditions(cmdLineArgs []string) error {
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(endpointMissingError)
	}
	return nil
}

// selectEndpoint ...
func selectEndpoint(cmd *cobra.Command, cmdLineArgs []string) {
	err := config.Config.SelectEndpoint(cmdLineArgs[0])
	if err != nil {
		if config.IsEndpointNotDefinedError(err) {
			fmt.Println(color.RedString("The endpoint given is not defined."))
			fmt.Println("Please use 'gsctl login <email> -e <endpoint>' to add a new endpoint first.")
			os.Exit(1)
		}
		fmt.Println(color.RedString("Error: " + err.Error()))
	} else {
		fmt.Println(color.GreenString("Endpoint selected: %s", config.Config.SelectedEndpoint))
	}
}
