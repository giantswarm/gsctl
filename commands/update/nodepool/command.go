// Package nodepool implements the "update nodepool" command.
package nodepool

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
	// Command is the cobra command for 'gsctl update nodepool'
	Command = &cobra.Command{
		Use:     "nodepool <cluster-id>/<nodepool-id>",
		Aliases: []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Modify node pool details",
		Long: `Change the name or the scaling settings of a node pool.

Examples:

  gsctl update nodepool f01r4/75rh1 --name "General purpose M5"

  gsctl update nodepool f01r4/75rh1 --nodes-min 10 --nodes-max 20

`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}
)

const (
	activityName = "update-nodepool"
)

func init() {
	initFlags()
}

// initFlags initializes flags in a re-usable way, so we can call it from multiple tests.
func initFlags() {
	Command.ResetFlags()
	Command.Flags().StringVarP(&flags.Name, "name", "n", "", "name or purpose description of the node pool")
	Command.Flags().Int64VarP(&flags.WorkersMin, "nodes-min", "", 0, "Minimum number of worker nodes for the node pool.")
	Command.Flags().Int64VarP(&flags.WorkersMax, "nodes-max", "", 0, "Maximum number of worker nodes for the node pool.")
}

// Arguments represents all the ways the user can influence the command.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterID         string
	Name              string
	NodePoolID        string
	ScalingMax        int64
	ScalingMin        int64
	UserProvidedToken string
}

func collectArguments(positionalArgs []string) (Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	parts := strings.Split(positionalArgs[0], "/")
	if len(parts) != 2 {
		return Arguments{}, microerror.Mask(errors.NodePoolIDMalformedError)
	}

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterID:         strings.TrimSpace(parts[0]),
		Name:              flags.Name,
		NodePoolID:        strings.TrimSpace(parts[1]),
		ScalingMax:        flags.WorkersMax,
		ScalingMin:        flags.WorkersMin,
		UserProvidedToken: flags.Token,
	}, nil
}

// result represents all information we get back from modifying a node pool.
type result struct {
	// nodePool contains all the node pool details as returned from the API.
	NodePool *models.V5GetNodePoolResponse
}

func verifyPreconditions(args Arguments) error {
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	} else if args.ClusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	} else if args.NodePoolID == "" {
		return microerror.Mask(errors.NodePoolIDMissingError)
	} else if args.ScalingMin == 0 && args.ScalingMax == 0 && args.Name == "" {
		return microerror.Mask(errors.NoOpError)
	} else if args.ScalingMin > args.ScalingMax && args.ScalingMax > 0 {
		return microerror.Mask(errors.WorkersMinMaxInvalidError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args, err := collectArguments(positionalArgs)
	if err == nil {
		err = verifyPreconditions(args)
	}

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	if errors.IsNodePoolIDMalformedError(err) {
		headline = "Bad format for Cluster ID/Node Pool ID argument"
		subtext = "Please provide cluster ID and node pool ID separated by a slash. See --help for examples."
	}

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

func updateNodePool(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	requestBody := &models.V5ModifyNodePoolRequest{}
	if args.Name != "" {
		requestBody.Name = args.Name
	}
	if args.ScalingMin != 0 || args.ScalingMax != 0 {
		requestBody.Scaling = &models.V5ModifyNodePoolRequestScaling{}
	}
	if args.ScalingMin != 0 {
		requestBody.Scaling.Min = args.ScalingMin
	}
	if args.ScalingMax != 0 {
		requestBody.Scaling.Max = args.ScalingMax
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.ModifyNodePool(args.ClusterID, args.NodePoolID, requestBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &result{
		NodePool: response.Payload,
	}

	return r, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args, _ := collectArguments(positionalArgs)

	r, err := updateNodePool(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

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

	fmt.Println(color.GreenString("Node pool '%s' (ID '%s') in cluster '%s' has been modified.", r.NodePool.Name, r.NodePool.ID, args.ClusterID))
}
