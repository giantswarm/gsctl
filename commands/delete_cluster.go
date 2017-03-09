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
		Use:   "cluster",
		Short: "Delete cluster",
		Long: `Deletes a Kubernetes cluster.

Caution: This will terminate all workloads an on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Examples:

	gsctl delete cluster --cluster=c7t2o

	gsctl delete cluster -c c7t2o`,
		PreRunE: checkDeleteCluster,
		Run:     deleteCluster,
	}
)

func init() {
	DeleteClusterCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// checks preconditions
func checkDeleteCluster(cmd *cobra.Command, args []string) error {
	// logged in?
	if config.Config.Token == "" && cmdToken == "" {
		s := color.RedString("You are not logged in.\n\n")
		return errors.New(s + "Use '" + config.ProgramName + " login' to login or '--auth-token' to pass a valid auth token.")
	}
	return nil
}

// interprets arguments/flags, eventually submits delete request
func deleteCluster(cmd *cobra.Command, args []string) {
	// perform API call
	authHeader := "giantswarm " + config.Config.Token
	if cmdToken != "" {
		// command line flag overwrites
		authHeader = "giantswarm " + cmdToken
	}
	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	responseBody, apiResponse, _ := client.DeleteCluster(authHeader, cmdClusterID, requestIDHeader, createClusterActivityName, cmdLine)

	// handle API result
	if responseBody.Code == "RESOURCE_DELETED" {
		fmt.Printf("The cluster with ID '%s' has been deleted\n\n", color.CyanString(cmdClusterID))
	} else if responseBody.Code == "RESOURCE_DELETION_STARTED" {
		fmt.Printf("The cluster with ID '%s' will be deleted soon\n\n", color.CyanString(cmdClusterID))
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
