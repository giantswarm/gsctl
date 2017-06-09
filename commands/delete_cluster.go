package commands

import (
	"errors"
	"fmt"
	"os"

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
	// auth token
	token string
	// verbosity
	verbose bool
}

func defaultDeleteClusterArguments() deleteClusterArguments {
	return deleteClusterArguments{
		apiEndpoint: cmdAPIEndpoint,
		clusterID:   cmdClusterID,
		force:       cmdForce,
		token:       cmdToken,
		verbose:     cmdVerbose,
	}
}

const (
	deleteClusterActivityName string = "delete-cluster"
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

	// force flag
	cmdForce bool

	// errors
	errClusterIDNotSpecified = "cluster ID not specified"
)

func init() {
	DeleteClusterCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	DeleteClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// validationOutput runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func deleteClusterValidationOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()

	err := validateDeleteClusterPreConditions(dca)
	if err != nil {
		var headline = ""
		var subtext = ""

		switch err.Error() {
		case "":
			return
		case errNotLoggedIn:
			headline = "You are not logged in."
			subtext = fmt.Sprintf("Use '%s login' to login or '--auth-token' to pass a valid auth token.", config.ProgramName)
		case errClusterIDNotSpecified:
			headline = "No cluster ID specified."
			subtext = "Please specify which cluster to delete using '-c' or '--cluster'."
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

// validatePreConditions checks preconditions and returns an error in case
func validateDeleteClusterPreConditions(args deleteClusterArguments) error {
	if args.clusterID == "" {
		return errors.New(errClusterIDNotSpecified)
	}
	if config.Config.Token == "" && args.token == "" {
		return errors.New(errNotLoggedIn)
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func deleteClusterExecutionOutput(cmd *cobra.Command, args []string) {
	dca := defaultDeleteClusterArguments()
	deleted, err := deleteCluster(dca)
	if err != nil {
		// show error
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

	// perform API call
	authHeader := "giantswarm " + config.Config.Token
	if args.token != "" {
		// command line flag overwrites
		authHeader = "giantswarm " + args.token
	}
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)
	responseBody, _, err := apiClient.DeleteCluster(authHeader, args.clusterID, requestIDHeader, createClusterActivityName, cmdLine)
	if err != nil {
		return false, err
	}

	// handle API result
	if responseBody.Code == "RESOURCE_DELETED" || responseBody.Code == "RESOURCE_DELETION_STARTED" {
		return true, nil
	}

	return false, fmt.Errorf("Error in API request to create cluster: %s (Code: %s)", responseBody.Message, responseBody.Code)
}
