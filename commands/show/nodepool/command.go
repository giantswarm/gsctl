// Package nodepool implements the 'show nodepool' command.
package nodepool

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/nodespec"
)

var (
	// ShowNodepoolCommand is the cobra command for 'gsctl show nodepool'
	ShowNodepoolCommand = &cobra.Command{
		Hidden:  true,
		Use:     "nodepool <cluster-id>/<nodepool-id>",
		Aliases: []string{"np"},
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

type arguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
	clusterID   string
	nodePoolID  string
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

func defaultArguments(positionalArgs []string) arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	parts := strings.Split(positionalArgs[0], "/")

	return arguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
		clusterID:   parts[0],
		nodePoolID:  parts[1],
	}
}

func verifyPreconditions(args arguments, positionalArgs []string) error {
	parsedArgs := defaultArguments(positionalArgs)
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
	args := defaultArguments(positionalArgs)
	err := verifyPreconditions(args, positionalArgs)
	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
}

// fetchNodePool collects all information we would want to display
// on a node pools of a cluster.
func fetchNodePool(args arguments) (*result, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	// create combined output data structure.
	res := &result{}

	response, err := clientV2.GetNodePool(args.clusterID, args.nodePoolID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res.nodePool = response.Payload

	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res.instanceTypeDetails, err = awsInfo.GetInstanceTypeDetails(res.nodePool.NodeSpec.Aws.InstanceType)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res.sumCPUs = res.nodePool.Status.NodesReady * int64(res.instanceTypeDetails.CPUCores)
	res.sumMemory = float64(res.nodePool.Status.NodesReady) * float64(res.instanceTypeDetails.MemorySizeGB)

	return res, nil

}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := defaultArguments(positionalArgs)
	data, err := fetchNodePool(args)
	if err != nil {
		errors.HandleCommonErrors(err)
	}

	table := []string{}

	table = append(table, color.YellowString("ID:")+"|"+data.nodePool.ID)
	table = append(table, color.YellowString("Name:")+"|"+data.nodePool.Name)
	table = append(table, color.YellowString("Node instance type:")+fmt.Sprintf("|%s - %d GB RAM, %d CPUs each",
		data.nodePool.NodeSpec.Aws.InstanceType, data.instanceTypeDetails.MemorySizeGB, data.instanceTypeDetails.CPUCores))
	table = append(table, color.YellowString("Availability zones:")+"|"+formatting.AvailabilityZonesList(data.nodePool.AvailabilityZones))
	table = append(table, color.YellowString("Node scaling:")+"|"+formatNodeScaling(data.nodePool.Scaling))
	table = append(table, color.YellowString("Nodes desired:")+fmt.Sprintf("|%d", data.nodePool.Status.Nodes))
	table = append(table, color.YellowString("Nodes in state Ready:")+fmt.Sprintf("|%d", data.nodePool.Status.NodesReady))
	table = append(table, color.YellowString("CPUs:")+fmt.Sprintf("|%d", data.nodePool.Status.NodesReady*int64(data.instanceTypeDetails.CPUCores)))
	table = append(table, color.YellowString("RAM:")+fmt.Sprintf("|%d GB", data.nodePool.Status.NodesReady*int64(data.instanceTypeDetails.MemorySizeGB)))

	fmt.Println(columnize.SimpleFormat(table))
}

func formatNodeScaling(scaling *models.V5GetNodePoolResponseScaling) string {
	if scaling.Min == scaling.Max {
		return fmt.Sprintf("Pinned to %d", scaling.Min)
	}

	return fmt.Sprintf("Autoscaling between %d and %d", scaling.Min, scaling.Max)
}
