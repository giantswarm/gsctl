// Package cluster implements the "update nodepool" command.
package cluster

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the cobra command for 'gsctl show nodepool'
	Command = &cobra.Command{
		Use: "cluster <cluster-id>",
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Modify cluster details",
		Long: `Change the name oif a cluster

Examples:

  gsctl update cluster f01r4 --name "Precious Production Cluster"

`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}
)

const (
	activityName = "update-cluster"
)

func init() {
	initFlags()
}

// initFlags initializes flags in a re-usable way, so we can call it from multiple tests.
func initFlags() {
	Command.ResetFlags()
	Command.Flags().StringVarP(&flags.Name, "name", "n", "", "new cluster name")
}

// Arguments represents all the ways the user can influence the command.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterID         string
	Name              string
	UserProvidedToken string
	Verbose           bool
}

func collectArguments(positionalArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterID:         strings.TrimSpace(positionalArgs[0]),
		Name:              flags.Name,
		UserProvidedToken: flags.Token,
		Verbose:           flags.Verbose,
	}
}

// result represents all information we get back from modifying a cluster.
type result struct {
	ClusterName string
}

func verifyPreconditions(args Arguments) error {
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	} else if args.ClusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	} else if args.Name == "" {
		return microerror.Mask(errors.NoOpError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)
	err := verifyPreconditions(args)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	// If there are specific errors to handle, add them here.
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

// updateCluster updates the cluster.
// It determines whether it is a v5 or v4 cluster and uses the appropriate mechanism.
func updateCluster(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	// First try v5.
	if args.Verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster via v5 API endpoint."))
	}
	_, errV5 := clientWrapper.GetClusterV5(args.ClusterID, auxParams)
	if errV5 == nil {
		requestBody := &models.V5ModifyClusterRequest{}
		if args.Name != "" {
			requestBody.Name = args.Name
		}

		if args.Verbose {
			fmt.Println(color.WhiteString("Sending cluster modification request to v5 endpoint."))
		}
		response, err := clientWrapper.ModifyClusterV5(args.ClusterID, requestBody, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r := &result{
			ClusterName: response.Payload.Name,
		}

		return r, nil
	}

	// Fallback: try v4.
	if args.Verbose {
		fmt.Println(color.WhiteString("No usable v5 response. Fetching details for cluster via v4 API endpoint."))
	}
	_, errV4 := clientWrapper.GetClusterV4(args.ClusterID, auxParams)
	if errV4 == nil {
		requestBody := &models.V4ModifyClusterRequest{}
		if args.Name != "" {
			requestBody.Name = args.Name
		}

		if args.Verbose {
			fmt.Println(color.WhiteString("Sending cluster modification request to v4 endpoint."))
		}
		response, err := clientWrapper.ModifyClusterV4(args.ClusterID, requestBody, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r := &result{
			ClusterName: response.Payload.Name,
		}

		return r, nil
	}

	// We return the last error here, representatively.
	return nil, microerror.Mask(errV4)
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)

	_, err := updateCluster(args)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		headline := ""
		subtext := ""

		switch {
		// If there are specific errors to handle, add them here.
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

	fmt.Println(color.GreenString("Cluster '%s' has been modified.", args.ClusterID))
}
