// Package nodepool implements the 'show nodepool' command.
package nodepool

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/clustercache"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
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

	headline := ""
	subtext := ""

	if errors.IsInvalidNodePoolIDArgument(err) {
		headline = "Invalid argument syntax"
		subtext = "Please give the cluster name or ID, followed by /, followed by the node pool ID."
	} else {
		headline = err.Error()
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
		}
	}

	return output, nil
}

func getOutputAWS(nodePool *models.V5GetNodePoolResponse) (string, error) {
	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return "", microerror.Mask(err)
	}

	instanceTypeDetails, err := awsInfo.GetInstanceTypeDetails(nodePool.NodeSpec.Aws.InstanceType)
	if nodespec.IsInstanceTypeNotFoundErr(err) {
		// We deliberately ignore "instance type not found", but respect all other errors.
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	var instanceTypes string
	{
		if len(nodePool.Status.InstanceTypes) > 0 {
			instanceTypes = strings.Join(nodePool.Status.InstanceTypes, ",")
		} else {
			instanceTypes = nodePool.NodeSpec.Aws.InstanceType
		}
	}

	var table []string
	{
		table = append(table, color.YellowString("ID:")+"|"+nodePool.ID)
		table = append(table, color.YellowString("Name:")+"|"+nodePool.Name)
		table = append(table, color.YellowString("Node instance types:")+"|"+formatInstanceType(instanceTypes, instanceTypeDetails))
		table = append(table, color.YellowString("Alike instances types:")+fmt.Sprintf("|%t", nodePool.NodeSpec.Aws.UseAlikeInstanceTypes))
		table = append(table, color.YellowString("Availability zones:")+"|"+formatting.AvailabilityZonesList(nodePool.AvailabilityZones))
		table = append(table, color.YellowString("On-demand base capacity:")+fmt.Sprintf("|%d", nodePool.NodeSpec.Aws.InstanceDistribution.OnDemandBaseCapacity))
		table = append(table, color.YellowString("Spot percentage above base capacity:")+fmt.Sprintf("|%d", 100-nodePool.NodeSpec.Aws.InstanceDistribution.OnDemandPercentageAboveBaseCapacity))
		table = append(table, color.YellowString("Node scaling:")+"|"+formatNodeScaling(nodePool.Scaling))
		table = append(table, color.YellowString("Nodes desired:")+fmt.Sprintf("|%d", nodePool.Status.Nodes))
		table = append(table, color.YellowString("Nodes in state Ready:")+fmt.Sprintf("|%d", nodePool.Status.NodesReady))
		table = append(table, color.YellowString("Spot instances:")+fmt.Sprintf("|%d", nodePool.Status.SpotInstances))
		table = append(table, color.YellowString("CPUs:")+"|"+formatCPUs(nodePool.Status.NodesReady, instanceTypeDetails))
		table = append(table, color.YellowString("RAM:")+"|"+formatRAM(nodePool.Status.NodesReady, instanceTypeDetails))
	}

	return columnize.SimpleFormat(table), nil
}

func formatInstanceType(instanceTypeName string, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%s - %d GB RAM, %d CPUs each",
			instanceTypeName,
			details.MemorySizeGB,
			details.CPUCores)
	}

	return fmt.Sprintf("%s %s", instanceTypeName, color.RedString("(no information available on this instance type)"))
}

func formatCPUs(numNodes int64, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%d", numNodes*int64(details.CPUCores))
	}

	return "n/a"
}

func formatRAM(numNodes int64, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%d GB", numNodes*int64(details.MemorySizeGB))
	}

	return "n/a"
}

func formatNodeScaling(scaling *models.V5GetNodePoolResponseScaling) string {
	if scaling.Min == scaling.Max {
		return fmt.Sprintf("Pinned to %d", scaling.Min)
	}

	return fmt.Sprintf("Autoscaling between %d and %d", scaling.Min, scaling.Max)
}
