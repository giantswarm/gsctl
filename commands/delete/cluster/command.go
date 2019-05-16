// Package cluster implements the 'delete cluster' sub-command.
package cluster

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
)

type deleteClusterArguments struct {
	// API endpoint
	apiEndpoint string
	// cluster ID to delete
	clusterID string
	// cluster ID passed via -c/--cluster argument
	legacyClusterID string
	// don't prompt
	force bool
	// auth scheme
	scheme string
	// auth token
	token string
	// verbosity
	verbose bool
}

func defaultArguments(positionalArgs []string) deleteClusterArguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	clusterID := ""
	if len(positionalArgs) > 0 {
		clusterID = positionalArgs[0]
	}

	return deleteClusterArguments{
		apiEndpoint:     endpoint,
		clusterID:       clusterID,
		force:           flags.CmdForce,
		legacyClusterID: flags.CmdClusterID,
		scheme:          scheme,
		token:           token,
		verbose:         flags.CmdVerbose,
	}
}

const (
	deleteClusterActivityName = "delete-cluster"
)

var (
	// Command performs the "delete cluster" function
	Command = &cobra.Command{
		Use:   "cluster",
		Short: "Delete cluster",
		Long: `Deletes a Kubernetes cluster.

Caution: This will terminate all workloads on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Example:

	gsctl delete cluster c7t2o`,
		PreRun: printValidation,
		Run:    printResult,
	}
)

func init() {
	Command.Flags().StringVarP(&flags.CmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	Command.Flags().BoolVarP(&flags.CmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")

	Command.Flags().MarkDeprecated("cluster", "You no longer need to pass the cluster ID with -c/--cluster. Use --help for details.")
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
		case errors.IsConflictingFlagsError(err):
			headline = "Conflicting flags/arguments"
			subtext = "Please specify the cluster to be used as a positional argument, avoid -c/--cluster."
			subtext += "See --help for details."
		case errors.IsClusterIDMissingError(err):
			headline = "No cluster ID specified"
			subtext = "See --help for usage details."
		case errors.IsCouldNotDeleteClusterError(err):
			headline = "The cluster could not be deleted."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"
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
func validatePreconditions(args deleteClusterArguments) error {
	if args.clusterID == "" && args.legacyClusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	if args.clusterID != "" && args.legacyClusterID != "" {
		return microerror.Mask(errors.ConflictingFlagsError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func printResult(cmd *cobra.Command, args []string) {
	dca := defaultArguments(args)
	deleted, err := deleteCluster(dca)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case errors.IsClusterNotFoundError(err):
			headline = "Cluster not found"
			subtext = "The cluster you tried to delete doesn't seem to exist. Check 'gsctl list clusters' to make sure."
		default:
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
		clusterID := dca.legacyClusterID
		if dca.clusterID != "" {
			clusterID = dca.clusterID
		}
		fmt.Println(color.GreenString("The cluster with ID '%s' will be deleted as soon as all workloads are terminated.", clusterID))
	} else {
		if dca.verbose {
			fmt.Println(color.GreenString("Aborted."))
		}
	}
}

// deleteCluster performs the cluster deletion API call
//
// The returned tuple contains:
// - bool: true if cluster will reall ybe deleted, false otherwise
// - error: The error that has occurred (or nil)
//
func deleteCluster(args deleteClusterArguments) (bool, error) {
	// Accept legacy cluster ID for a while, but real one takes precedence.
	clusterID := args.legacyClusterID
	if args.clusterID != "" {
		clusterID = args.clusterID
	}

	// confirmation
	if !args.force {
		confirmed := confirm.Ask("Do you really want to delete cluster '" + clusterID + "'?")
		if !confirmed {
			return false, nil
		}
	}

	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return false, microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = deleteClusterActivityName

	// perform API call
	_, err = clientV2.DeleteCluster(clusterID, auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return false, microerror.Mask(errors.AccessForbiddenError)
			} else if clientErr.HTTPStatusCode == http.StatusNotFound {
				return false, microerror.Mask(errors.ClusterNotFoundError)
			}
		}

		return false, microerror.Maskf(errors.CouldNotDeleteClusterError, err.Error())
	}

	return true, nil
}
