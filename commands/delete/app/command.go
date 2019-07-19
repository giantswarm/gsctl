// Package app implements the 'delete app' sub-command.
package app

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
)

type deleteAppArguments struct {
	// API endpoint
	apiEndpoint string
	// app to delete
	appName string
	// cluster ID
	clusterID string
	// don't prompt
	force bool
	// auth scheme
	scheme string
	// auth token
	token string
	// verbosity
	verbose bool
}

func defaultArguments(positionalArgs []string) deleteAppArguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	appName := ""
	if len(positionalArgs) > 0 {
		appName = positionalArgs[0]
	}

	return deleteAppArguments{
		apiEndpoint: endpoint,
		appName:     appName,
		clusterID:   flags.CmdClusterID,
		force:       flags.CmdForce,
		scheme:      scheme,
		token:       token,
		verbose:     flags.CmdVerbose,
	}
}

const (
	deleteAppActivityName = "delete-app"
)

var (
	// Command performs the "delete app" function
	Command = &cobra.Command{
		Use:   "app",
		Short: "Delete app",
		Long: `Deletes an app on a tenant cluster.

Example:

	gsctl delete app my-grafana -c c7t2o`,
		PreRun: printValidation,
		Run:    printResult,
	}
)

func init() {
	Command.Flags().StringVarP(&flags.CmdClusterID, "cluster", "c", "", "ID of the tenant cluster")
	Command.Flags().BoolVarP(&flags.CmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
}

// printValidation runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func printValidation(cmd *cobra.Command, args []string) {
	dca := defaultArguments(args)

	err := validatePreconditions(dca)
	if err != nil {
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case errors.IsAppNameMissingError(err):
			headline = "No app name specified"
			subtext = "See --help for usage details."
		case errors.IsClusterIDMissingError(err):
			headline = "No cluster ID specified"
			subtext = "See --help for usage details."
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

// validatePreconditions checks preconditions and returns an error in case
func validatePreconditions(args deleteAppArguments) error {
	if args.clusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	if args.appName == "" {
		return microerror.Mask(errors.AppNameMissingError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := defaultArguments(positionalArgs)
	deleted, err := deleteApp(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		var headline = ""
		var subtext = ""

		if clientErr, ok := err.(*clienterror.APIError); ok {
			headline = clientErr.ErrorMessage
			subtext = clientErr.ErrorDetails
		} else {
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	// non-error output
	if deleted {
		fmt.Println(color.GreenString("App %q on cluster %q will be deleted as soon as possible.", args.appName, args.clusterID))
	} else {
		if args.verbose {
			fmt.Println(color.GreenString("Aborted."))
		}
	}
}

// deleteApp performs the app deletion API call
//
// The returned tuple contains:
// - bool: true if the app will really be deleted, false otherwise
// - error: The error that has occurred (or nil)
//
func deleteApp(args deleteAppArguments) (bool, error) {
	// confirmation
	if !args.force {
		confirmed := confirm.Ask(fmt.Sprintf("Do you really want to delete app %q on cluster %q?", args.appName, args.clusterID))
		if !confirmed {
			return false, nil
		}
	}

	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return false, microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = deleteAppActivityName

	// perform API call
	_, err = clientV2.DeleteApp(args.clusterID, args.appName, auxParams)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}
