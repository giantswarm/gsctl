package commands

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client"
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

Caution: This will terminate all workloads on the cluster. Data stored on the
worker nodes will be lost. There is no way to undo this.

Example:

	gsctl delete cluster -c c7t2o`,
		PreRunE: checkDeleteCluster,
		Run:     deleteCluster,
	}

	// force flag
	cmdForce bool
)

func init() {
	DeleteClusterCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to delete")
	DeleteClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
	DeleteCommand.AddCommand(DeleteClusterCommand)
}

// checks preconditions
func checkDeleteCluster(cmd *cobra.Command, args []string) error {
	if cmdClusterID == "" {
		return errors.New(color.RedString("Please select a cluster to delete"))
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

	// confirmation
	if cmdForce == false {
		confirmed := askForConfirmation("Do you really want to delete cluster '" + cmdClusterID + "'?")
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
	clientConfig := client.Configuration{
		Endpoint:  cmdAPIEndpoint,
		Timeout:   60 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)
	responseBody, apiResponse, _ := apiClient.DeleteCluster(authHeader, cmdClusterID, requestIDHeader, createClusterActivityName, cmdLine)

	// handle API result
	if responseBody.Code == "RESOURCE_DELETED" || responseBody.Code == "RESOURCE_DELETION_STARTED" {
		fmt.Println(color.GreenString("The cluster with ID '%s' will be deleted as soon as all workloads are terminated.", cmdClusterID))
	} else {
		fmt.Println()
		fmt.Println(color.RedString("Could not delete cluster"))
		fmt.Printf("Error message: %s\n", responseBody.Message)
		fmt.Printf("Error code: %s\n", responseBody.Code)
		fmt.Println(fmt.Sprintf("Raw response body:\n%v", string(apiResponse.Payload)))
		fmt.Println("Please contact Giant Swarm via support@giantswarm.io in case you need any help.")
		os.Exit(1)
	}
}
