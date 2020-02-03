// Package endpoint implements the 'delete endpoint' sub-command.
package endpoint

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
)

// Arguments represents all argument that can be passed to our
// business function.
type Arguments struct {
	// API endpoint to delete
	APIEndpoint string
	// Don't prompt
	Force bool
	// Verbosity
	Verbose bool
}

func collectArguments(positionalArgs []string) Arguments {
	var endpointToDelete string
	if len(positionalArgs) > 0 {
		endpointToDelete = positionalArgs[0]
	}

	return Arguments{
		APIEndpoint: endpointToDelete,
		Force:       flags.Force,
		Verbose:     flags.Verbose,
	}
}

var (
	// Command performs the "delete endpoint" function
	Command = &cobra.Command{
		Use:   "endpoint",
		Short: "Delete endpoint",
		Long: `Deletes an API endpoint.

Caution: This will remove the API endpoint from the existing configuration. To use it again, you will have to re-authenticate. There is no way to undo this.

Example:

	gsctl delete endpoint https://api.gigantic.io`,
		PreRun: printValidation,
		Run:    printResult,
	}

	arguments Arguments
)

func init() {
	Command.Flags().BoolVarP(&flags.Force, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
}

// printValidation runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func printValidation(cmd *cobra.Command, args []string) {
	arguments = collectArguments(args)

	err := validatePreconditions(arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case errors.IsEndpointMissingError(err):
			headline = "No API endpoint specified"
			subtext = "See --help for usage details."
		case errors.IsCouldNotDeleteEndpointError(err):
			headline = "The API endpoint could not be deleted."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += "Sorry for the inconvenience!"
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

// validatePreconditions checks preconditions and returns
// an error in case they are invalid
func validatePreconditions(args Arguments) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}

	return nil
}

// interprets arguments/flags, eventually submits delete request
func printResult(cmd *cobra.Command, args []string) {
	deleted, err := deleteEndpoint(arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case errors.IsEndpointNotFoundError(err):
			headline = "API Endpoint not found"
			subtext = "The API endpoint you are trying to delete does not exist. Check 'gsctl list endpoints' to make sure"
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}

		os.Exit(1)
	}

	// Non-error output
	if deleted {
		fmt.Println(color.GreenString("The API endpoint '%s' deleted successfully.", arguments.APIEndpoint))
	} else {
		if arguments.Verbose {
			fmt.Println(color.GreenString("Aborted."))
		}
	}
}

// deleteEndpoint performs the endpoint deletion operation
//
// The returned tuple contains:
// - bool: true if endpoint is deleted, false otherwise
// - error: The error that has occurred (or nil)
func deleteEndpoint(args Arguments) (bool, error) {
	// Confirmation
	if !args.Force {
		confirmed := confirm.AskStrict("Do you really want to delete API endpoint '"+args.APIEndpoint+"'? Please type the endpoint name to confirm", args.APIEndpoint)
		if !confirmed {
			return false, nil
		}
	}

	// Delete Endpoint
	err := config.Config.DeleteEndpoint(args.APIEndpoint)
	if err != nil {
		if config.IsEndpointNotDefinedError(err) {
			return false, microerror.Mask(errors.EndpointNotFoundError)
		}

		return false, microerror.Maskf(errors.CouldNotDeleteEndpointError, err.Error())
	}

	return true, nil
}
