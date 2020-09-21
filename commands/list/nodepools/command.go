// Package nodepools implements the 'list nodepools' sub-command.
package nodepools

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
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
	// Command performs the "list nodepools" function
	Command = &cobra.Command{
		Use:     "nodepools <cluster-name/cluster-id>",
		Aliases: []string{"nps", "np"},

		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "List node pools",
		Long: `Prints a list of the node pools of a cluster.

The result will be a table of all node pools of a specific cluster with the following details in
columns:

	ID:              Node pool identifier (unique within the cluster)
	NAME:            Name specified for the node pool, usually indicating the purpose
	AZ:              Availability zone letters used by the node pool, separated by comma
	INSTANCE TYPES:  EC2 instance types used for worker nodes
	ALIKE:           If similar instance types are allowed in your node pool. This list is maintained by Giant Swarm at the moment. Eg if you select m5.xlarge then the node pool can fall back on m4.xlarge too
	ON-DEMAND BASE:  Number of on-demand instances that this node pool needs to have until spot instances are used.
	SPOT PERCENTAGE: Percentage of spot instances used once the on-demand base capacity is fullfilled. A number of 40 means that 60% will be on-demand and 40% will be spot instances.
	NODES MIN/MAX:   The minimum and maximum number of worker nodes in this pool
	NODES DESIRED:   Current desired number of nodes as determined by the autoscaler
	NODES READY:     Number of nodes that are in the Ready state in kubernetes
	SPOT INSTANCES:  Number of spot instances in this node pool
	CPUS:            Sum of CPU cores in nodes that are in state Ready
	RAM (GB):        Sum of memory in GB of all nodes that are in state Ready

To see all available details for a cluster, use 'gsctl show nodepool <cluster-id>/<nodepool-id>'.

To list all clusters you have access to, use 'gsctl list clusters'.`,
		PreRun: printValidation,
		Run:    printResult,
	}

	arguments Arguments
)

const activityName = "list-nodepools"

func init() {
	initFlags()
}

func initFlags() {
	Command.Flags().StringVarP(&flags.OutputFormat, "output", "o", formatting.OutputFormatTable, fmt.Sprintf("Use '%s' for JSON output. Defaults to human-friendly table output.", formatting.OutputFormatJSON))
}

type Arguments struct {
	apiEndpoint       string
	authToken         string
	clusterNameOrID   string
	outputFormat      string
	scheme            string
	userProvidedToken string
	verbose           bool
}

// collectArguments creates arguments based on command line flags and config.
func collectArguments(cmdLineArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		clusterNameOrID:   cmdLineArgs[0],
		outputFormat:      flags.OutputFormat,
		scheme:            scheme,
		userProvidedToken: flags.Token,
		verbose:           flags.Verbose,
	}
}

func verifyPreconditions(args Arguments, positionalArgs []string) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.outputFormat != formatting.OutputFormatJSON && args.outputFormat != formatting.OutputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, "Output format '%s' is unknown", args.outputFormat)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	arguments = collectArguments(positionalArgs)
	err := verifyPreconditions(arguments, positionalArgs)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}
}

// fetchNodePools collects all information we would want to display
// on the node pools of a cluster.
func fetchNodePools(args Arguments) ([]*models.V5GetNodePoolsResponseItems, error) {
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

	response, err := clientWrapper.GetNodePools(clusterID, auxParams)
	if err != nil {
		if errors.IsClusterNotFoundError(err) {
			// Check if there is a v4 cluster of this name/ID, to provide a specific error for this case.
			if args.verbose {
				fmt.Println(color.WhiteString("Couldn't find a node pools (v5) cluster with name/ID %s. Checking v4.", args.clusterNameOrID))
			}

			_, err := clientWrapper.GetClusterV4(clusterID, auxParams)
			if err == nil {
				return nil, microerror.Mask(errors.ClusterDoesNotSupportNodePoolsError)
			}
		}

		return nil, microerror.Mask(err)
	}

	// sort node pools by ID
	sort.Slice(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	return response.Payload, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	nodePools, err := fetchNodePools(arguments)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}

	if len(nodePools) == 0 {
		fmt.Println(color.YellowString("This cluster has no node pools"))
		return
	}

	output, err := getOutput(nodePools, arguments.outputFormat)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}
	// Display output.
	fmt.Println(output)
}

func formatNodesReady(nodes, nodesReady int64) string {
	if nodes == nodesReady {
		return strconv.FormatInt(nodesReady, 10)
	}

	return color.YellowString(strconv.FormatInt(nodesReady, 10))
}

func getOutput(nps []*models.V5GetNodePoolsResponseItems, outputFormat string) (string, error) {
	if len(nps) < 0 {
		return "", nil
	}

	if outputFormat == formatting.OutputFormatJSON {
		outputBytes, err := json.MarshalIndent(nps, formatting.OutputJSONPrefix, formatting.OutputJSONIndent)
		if err != nil {
			return "", microerror.Mask(err)
		}

		return string(outputBytes), nil
	}

	var output string
	var err error
	np := nps[0]

	if np.NodeSpec.Aws != nil && np.NodeSpec.Azure == nil {
		output, err = getOutputAWS(nps)
		if err != nil {
			return "", microerror.Mask(err)
		}
	} else if np.NodeSpec.Azure != nil && np.NodeSpec.Aws == nil {
		output, err = getOutputAzure(nps)
		if err != nil {
			return "", microerror.Mask(err)
		}
	} else {
		return "", microerror.Mask(errors.ClusterDoesNotSupportNodePoolsError)
	}

	return output, nil
}

func getOutputAWS(nps []*models.V5GetNodePoolsResponseItems) (string, error) {
	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return "", microerror.Mask(err)
	}

	headers := []string{
		color.CyanString("ID"),
		color.CyanString("NAME"),
		color.CyanString("AZ"),
		color.CyanString("INSTANCE TYPE"),
		color.CyanString("ALIKE"),
		color.CyanString("ON-DEMAND BASE"),
		color.CyanString("SPOT PERCENTAGE"),
		color.CyanString("NODES MIN/MAX"),
		color.CyanString("NODES DESIRED"),
		color.CyanString("NODES READY"),
		color.CyanString("SPOT INSTANCES"),
		color.CyanString("CPUS"),
		color.CyanString("RAM (GB)"),
	}

	table := make([]string, 0, len(nps)+1)
	table = append(table, strings.Join(headers, "|"))

	for _, np := range nps {
		it, err := awsInfo.GetInstanceTypeDetails(np.NodeSpec.Aws.InstanceType)
		if nodespec.IsInstanceTypeNotFoundErr(err) {
			// We deliberately ignore "instance type not found", but respect all other errors.
		} else if err != nil {
			return "", microerror.Mask(err)
		}

		var sumCPUs string
		{
			if it == nil {
				sumCPUs = "n/a"
			} else {
				totalCPUs := np.Status.NodesReady * int64(it.CPUCores)
				sumCPUs = strconv.FormatInt(totalCPUs, 10)
			}
		}

		var sumMemory string
		{
			if it == nil {
				sumMemory = "n/a"
			} else {
				totalMemory := float64(np.Status.NodesReady) * float64(it.MemorySizeGB)
				sumMemory = strconv.FormatFloat(totalMemory, 'f', 1, 64)
			}
		}

		var instanceTypes string
		{
			if len(np.Status.InstanceTypes) > 0 {
				instanceTypes = strings.Join(np.Status.InstanceTypes, ",")
			} else {
				instanceTypes = np.NodeSpec.Aws.InstanceType
			}
		}

		table = append(table, strings.Join([]string{
			np.ID,
			np.Name,
			formatting.AvailabilityZonesList(np.AvailabilityZones),
			instanceTypes,
			fmt.Sprintf("%t", np.NodeSpec.Aws.UseAlikeInstanceTypes),
			strconv.FormatInt(np.NodeSpec.Aws.InstanceDistribution.OnDemandBaseCapacity, 10),
			strconv.FormatInt(100-np.NodeSpec.Aws.InstanceDistribution.OnDemandPercentageAboveBaseCapacity, 10),
			strconv.FormatInt(np.Scaling.Min, 10) + "/" + strconv.FormatInt(np.Scaling.Max, 10),
			strconv.FormatInt(np.Status.Nodes, 10),
			formatNodesReady(np.Status.Nodes, np.Status.NodesReady),
			strconv.FormatInt(np.Status.SpotInstances, 10),
			sumCPUs,
			sumMemory,
		}, "|"))
	}

	colConfig := columnize.DefaultConfig()
	colConfig.ColumnSpec = []*columnize.ColumnSpecification{
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
	}

	return columnize.Format(table, colConfig), nil
}

func getOutputAzure(nps []*models.V5GetNodePoolsResponseItems) (string, error) {
	azureInfo, err := nodespec.NewAzureProvider()
	if err != nil {
		return "", microerror.Mask(err)
	}

	headers := []string{
		color.CyanString("ID"),
		color.CyanString("NAME"),
		color.CyanString("AZ"),
		color.CyanString("VM SIZE"),
		color.CyanString("NODES MIN/MAX"),
		color.CyanString("NODES DESIRED"),
		color.CyanString("NODES READY"),
		color.CyanString("CPUS"),
		color.CyanString("RAM (GB)"),
	}

	table := make([]string, 0, len(nps)+1)
	table = append(table, strings.Join(headers, "|"))

	for _, np := range nps {
		vmSize, err := azureInfo.GetVMSizeDetails(np.NodeSpec.Azure.VMSize)
		if nodespec.IsVMSizeNotFoundErr(err) {
			// We deliberately ignore "vm size not found", but respect all other errors.
		} else if err != nil {
			return "", microerror.Mask(err)
		}

		var sumCPUs string
		{
			if vmSize == nil {
				sumCPUs = "n/a"
			} else {
				totalCPUs := np.Status.NodesReady * vmSize.NumberOfCores
				sumCPUs = strconv.FormatInt(totalCPUs, 10)
			}
		}

		var sumMemory string
		{
			if vmSize == nil {
				sumMemory = "n/a"
			} else {
				totalMemory := float64(np.Status.NodesReady) * vmSize.MemoryInMB / 1000
				sumMemory = strconv.FormatFloat(totalMemory, 'f', 1, 64)
			}
		}

		var vmSizes string
		{
			if len(np.Status.InstanceTypes) > 0 {
				vmSizes = strings.Join(np.Status.InstanceTypes, ",")
			} else {
				vmSizes = np.NodeSpec.Azure.VMSize
			}
		}

		table = append(table, strings.Join([]string{
			np.ID,
			np.Name,
			formatting.AvailabilityZonesList(np.AvailabilityZones),
			vmSizes,
			strconv.FormatInt(np.Scaling.Min, 10) + "/" + strconv.FormatInt(np.Scaling.Max, 10),
			strconv.FormatInt(np.Status.Nodes, 10),
			formatNodesReady(np.Status.Nodes, np.Status.NodesReady),
			sumCPUs,
			sumMemory,
		}, "|"))
	}

	colConfig := columnize.DefaultConfig()
	colConfig.ColumnSpec = []*columnize.ColumnSpecification{
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignLeft},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
		&columnize.ColumnSpecification{Alignment: columnize.AlignRight},
	}

	return columnize.Format(table, colConfig), nil
}

func handleError(err error) {
	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	case errors.IsClusterDoesNotSupportNodePools(err):
		headline = "This cluster does not support node pools."
		subtext = "Node pools cannot be listed for this cluster. Please use 'gsctl show cluster' to get information on worker nodes."
	default:
		headline = err.Error()
	}

	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
}
