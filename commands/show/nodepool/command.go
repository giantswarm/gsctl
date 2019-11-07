// Package nodepool implements the 'show nodepool' command.
package nodepool

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

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
		Hidden:                true,
		Use:                   "nodepool <cluster-id>/<nodepool-id>",
		Aliases:               []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Show node pool details",
		Long: `Display details of a node pool.

Examples:

  gsctl show nodepool f01r4/75rh1
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
	clusterID         string
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

func collectArguments(positionalArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	parts := strings.Split(positionalArgs[0], "/")

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		clusterID:         parts[0],
		nodePoolID:        parts[1],
		userProvidedToken: flags.Token,
	}
}

func verifyPreconditions(args Arguments, positionalArgs []string) error {
	parsedArgs := collectArguments(positionalArgs)
	if config.Config.Token == "" && parsedArgs.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	if parsedArgs.clusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	if parsedArgs.nodePoolID == "" {
		return microerror.Mask(errors.NodePoolIDMissingError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)
	err := verifyPreconditions(args, positionalArgs)
	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)
}

// fetchNodePool collects all information we would want to display
// on a node pools of a cluster.
func fetchNodePool(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	// create combined output data structure.
	res := &result{}

	response, err := clientWrapper.GetNodePool(args.clusterID, args.nodePoolID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res.nodePool = response.Payload

	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res.instanceTypeDetails, err = awsInfo.GetInstanceTypeDetails(res.nodePool.NodeSpec.Aws.InstanceType)
	if nodespec.IsInstanceTypeNotFoundErr(err) {
		// We deliberately ignore "instance type not found", but respect all other errors.
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		res.sumCPUs = res.nodePool.Status.NodesReady * int64(res.instanceTypeDetails.CPUCores)
		res.sumMemory = float64(res.nodePool.Status.NodesReady) * float64(res.instanceTypeDetails.MemorySizeGB)
	}

	return res, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)
	data, err := fetchNodePool(args)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)
	}

	table := []string{}

	table = append(table, color.YellowString("ID:")+"|"+data.nodePool.ID)
	table = append(table, color.YellowString("Name:")+"|"+data.nodePool.Name)
	table = append(table, color.YellowString("Node instance type:")+"|"+formatInstanceType(data.nodePool.NodeSpec.Aws.InstanceType, data.instanceTypeDetails))
	table = append(table, color.YellowString("Availability zones:")+"|"+formatting.AvailabilityZonesList(data.nodePool.AvailabilityZones))
	table = append(table, color.YellowString("Node scaling:")+"|"+formatNodeScaling(data.nodePool.Scaling))
	table = append(table, color.YellowString("Nodes desired:")+fmt.Sprintf("|%d", data.nodePool.Status.Nodes))
	table = append(table, color.YellowString("Nodes in state Ready:")+fmt.Sprintf("|%d", data.nodePool.Status.NodesReady))
	table = append(table, color.YellowString("CPUs:")+"|"+formatCPUs(data.nodePool.Status.NodesReady, data.instanceTypeDetails))
	table = append(table, color.YellowString("RAM:")+"|"+formatRAM(data.nodePool.Status.NodesReady, data.instanceTypeDetails))

	fmt.Println(columnize.SimpleFormat(table))
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
