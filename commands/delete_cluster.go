package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

const (
	deleteClusterActivityName string = "delete-cluster"
)

var (

	// DeleteClusterCommand performs the "delete cluster" function
	DeleteClusterCommand = &cobra.Command{
		Use:   "cluster <cluster_id>",
		Short: "Delete cluster",
		Long: `Deletes a Kubernetes cluster.

Caution: This will terminate all workloads an on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Examples:

	gsctl delete cluster c7t2o`,
		PreRunE: checkDeleteCluster,
		Run:     deleteCluster,
	}

	// force flag
	cmdForce bool
)

func init() {
	DeleteClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required.")
	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// checks preconditions
func checkDeleteCluster(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New(color.RedString("The cluster_id argument is required"))
	}

	// logged in?
	if config.Config.Token == "" && cmdToken == "" {
		s := color.RedString("You are not logged in.\n\n")
		return errors.New(s + "Use '" + config.ProgramName + " login' to login or '--auth-token' to pass a valid auth token.")
	}

	return nil
}

// interprets arguments/flags, eventually submits delete request
func deleteCluster(cmd *cobra.Command, args []string) {
	var clusterID = args[0]

	// confirmation
	if cmdForce == false {
		confirmed := askForConfirmation("Do you really want to delete cluster '" + clusterID + "'?")
		if !confirmed {
			if cmdVerbose {
				fmt.Println("Cluster not deleted")
			}
			os.Exit(0)
		}
	}

	// perform API call
	authHeader := "giantswarm " + config.Config.Token
	if cmdToken != "" {
		// command line flag overwrites
		authHeader = "giantswarm " + cmdToken
	}
	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	responseBody, apiResponse, _ := client.DeleteCluster(authHeader, clusterID, requestIDHeader, createClusterActivityName, cmdLine)

	// handle API result
	if responseBody.Code == "RESOURCE_DELETED" {
		fmt.Printf("The cluster with ID '%s' has been deleted\n", color.CyanString(clusterID))
	} else if responseBody.Code == "RESOURCE_DELETION_STARTED" {
		fmt.Printf("The cluster with ID '%s' will be deleted soon\n", color.CyanString(clusterID))
	} else {
		fmt.Println()
		fmt.Println(color.RedString("Could not delete cluster"))
		fmt.Printf("Error message: %s\n", responseBody.Message)
		fmt.Printf("Error code: %d\n", responseBody.Code)
		fmt.Println(fmt.Sprintf("Raw response body:\n%v", string(apiResponse.Payload)))
		fmt.Println("Please contact Giant Swarm via support@giantswarm.io in case you need any help.")
		os.Exit(1)
	}
}
