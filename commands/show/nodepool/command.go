// Package nodepool implements the 'show nodepool' command.
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

	"github.com/giantswarm/gsctl/clustercache"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/nodespec"
)

var (
	// ShowNodepoolCommand is the cobra command for 'gsctl show nodepool'
	ShowNodepoolCommand = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "nodepool <cluster-name/cluster-id>/<nodepool-id>",
		Aliases:               []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Show node pool details",
		Long: `Display details of a node pool.

Examples:

  gsctl show nodepool f01r4/75rh1
  gsctl show nodepool "Cluster name"/75rh1
`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}
)

const (
	activityName = "show-nodepool"
)

type Arguments struct {
	apiEndpoint       string
	authToken         string
	clusterNameOrID   string
	nodePoolID        string
	userProvidedToken string
}

// result represents all information we want to collect about one node pool.
type result struct {
	// nodePool contains all the node pool details as returned from the API.
	nodePool *models.V5GetNodePoolResponse
	// instanceTypeDetails contains details on the instance type.
	instanceTypeDetails *nodespec.InstanceType
	sumCPUs             int64
	sumMemory           float64
}

func collectArguments(positionalArgs []string) (*Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	parts := strings.Split(positionalArgs[0], "/")

	if len(parts) < 2 {
		return nil, microerror.Maskf(errors.InvalidNodePoolIDArgumentError, "Please specify the node pool as <cluster-name/cluster-id>/<nodepool-id>. Use --help for details.")
	}

	return &Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		clusterNameOrID:   parts[0],
		nodePoolID:        parts[1],
		userProvidedToken: flags.Token,
	}, nil
}

func verifyPreconditions(args *Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	if args.clusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	}
	if args.nodePoolID == "" {
		return microerror.Mask(errors.NodePoolIDMissingError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args, err := collectArguments(positionalArgs)
	if err == nil {
		err = verifyPreconditions(args)
		if err == nil {
			return
		}
	}

	handleError(err)
	os.Exit(1)
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	output, err := getOutput(positionalArgs)
	if err != nil {
		handleError(microerror.Mask(err))
		os.Exit(1)
	}

	fmt.Println(output)
}

func handleError(err error) {
	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	var (
		headline string
		subtext  string
	)
	{
		switch {
		case errors.IsClusterDoesNotSupportNodePools(err):
			headline = "This cluster does not support node pools."
			subtext = "Node pools cannot be listed for this cluster. Please use 'gsctl show cluster' to get information on worker nodes."

		case errors.IsInvalidNodePoolIDArgument(err):
			headline = "Invalid argument syntax"
			subtext = "Please give the cluster name or ID, followed by /, followed by the node pool ID."

		default:
			headline = err.Error()
		}
	}

	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
}

// fetchNodePool collects all information we would want to display
// on a node pools of a cluster.
func fetchNodePool(args *Arguments) (*models.V5GetNodePoolResponse, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.apiEndpoint, args.clusterNameOrID, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetNodePool(clusterID, args.nodePoolID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}

func getOutput(positionalArgs []string) (string, error) {
	args, err := collectArguments(positionalArgs)
	if err != nil {
		return "", microerror.Mask(err)
	}

	nodePool, err := fetchNodePool(args)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var output string
	{
		switch {
		case nodePool.NodeSpec.Aws != nil:
			output, err = getOutputAWS(nodePool)
			if err != nil {
				return "", microerror.Mask(err)
			}

		case nodePool.NodeSpec.Azure != nil:
			output, err = getOutputAzure(nodePool)
			if err != nil {
				return "", microerror.Mask(err)
			}

		default:
			return "", microerror.Mask(errors.ClusterDoesNotSupportNodePoolsError)
		}
	}

	return output, nil
}
