package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/giantswarm/microerror"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

type deleteClusterArguments struct {
	// API endpoint
	apiEndpoint string
	// cluster ID to delete
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

func defaultDeleteClusterArguments() deleteClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return deleteClusterArguments{
		apiEndpoint: endpoint,
		clusterID:   cmdClusterID,
		force:       cmdForce,
		scheme:      scheme,
		token:       token,
		verbose:     cmdVerbose,
	}
}

const (
	deleteClusterActivityName = "delete-cluster"
)

var (

	// DeleteClusterCommand performs the "delete cluster" function
	DeleteClusterCommand = &cobra.Command{
		Use:   "cluster",
		Short: "Delete cluster",
		Long: `Deletes a Kubernetes cluster.

Caution: This will terminate all workloads on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Example:

	gsctl delete cluster -c c7t2o`,
		PreRun: deleteClusterValidationOutput,
		Run:    deleteClusterExecutionOutput,
	}
)

func init() {
	DeleteClusterCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	DeleteClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")

	DeleteClusterCommand.MarkFlagRequired("cluster")

	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// deleteClusterValidationOutput runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func deleteClusterValidationOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()

	err := validateDeleteClusterPreConditions(dca)
	if err != nil {
		handleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case err.Error() == "":
			return
		case IsCouldNotDeleteClusterError(err):
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

// validateDeleteClusterPreConditions checks preconditions and returns an error in case
func validateDeleteClusterPreConditions(args deleteClusterArguments) error {
	if args.clusterID == "" {
		return microerror.Mask(clusterIDMissingError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func deleteClusterExecutionOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()
	deleted, err := deleteCluster(dca)
	if err != nil {
		handleCommonErrors(err)

		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	// non-error output
	if deleted {
		fmt.Println(color.GreenString("The cluster with ID '%s' will be deleted as soon as all workloads are terminated.", dca.clusterID))
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
	// confirmation
	if !args.force {
		confirmed := askForConfirmation("Do you really want to delete cluster '" + args.clusterID + "'?")
		if !confirmed {
			return false, nil
		}
	}

	// prepare API client
	authHeader := args.scheme + " " + args.token
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return false, microerror.Mask(couldNotCreateClientError)
	}

	// perform API call
	responseBody, rawResponse, err := apiClient.DeleteCluster(authHeader,
		args.clusterID, requestIDHeader, createClusterActivityName, cmdLine)
	if err != nil {
		if rawResponse.Response != nil && rawResponse.Response.StatusCode == http.StatusForbidden {
			return false, microerror.Mask(accessForbiddenError)
		}
		return false, microerror.Mask(err)
	}

	// handle API result
	if responseBody.Code == "RESOURCE_DELETED" || responseBody.Code == "RESOURCE_DELETION_STARTED" {
		return true, nil
	}

	return false, microerror.Maskf(couldNotDeleteClusterError,
		fmt.Sprintf("Error in API request to create cluster: %s (Code: %s)",
			responseBody.Message, responseBody.Code))
}
