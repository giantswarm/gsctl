// Package nodepools implements the 'list nodepools' sub-command.
package nodepools

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/nodespec"
)

var (
	// Command performs the "list nodepools" function
	Command = &cobra.Command{
		Use:     "nodepools <cluster-id>",
		Aliases: []string{"nps", "np"},

		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "List node pools",
		Long: `Prints a list of the node pools of a cluster.

The result will be a table of all node pools of a specific cluster with the following details in
columns:

	ID:             Node pool identifier (unique within the cluster)
	NAME:           Name specified for the node pool, usually indicating the purpose
	AZ:             Availability zone letters used by the node pool, separated by comma
	INSTANCE TYPE:  EC2 instance type used for worker nodes
	NODES MIN/MAX:  The minimum and maximum number of worker nodes in this pool
	NODES DESIRED:  Current desired number of nodes as determined by the autoscaler
	NODES READY:    Number of nodes that are in the Ready state in kubernetes
	CPUS:           Sum of CPU cores in nodes that are in state Ready
	RAM (GB):       Sum of memory in GB of all nodes that are in state Ready

To see all available details for a cluster, use 'gsctl show nodepool <cluster-id>/<nodepool-id>'.

To list all clusters you have access to, use 'gsctl list clusters'.`,
		PreRun: printValidation,
		Run:    printResult,
	}
)

const activityName = "list-nodepools"

type Arguments struct {
	apiEndpoint       string
	authToken         string
	clusterID         string
	scheme            string
	userProvidedToken string
}

// resultRow represents one nope pool row as returned by fetchNodePools.
type resultRow struct {
	// nodePool contains all the node pool details as returned from the API.
	nodePool *models.V5GetNodePoolsResponseItems
	// instanceTypeDetails contains details on the instance type.
	instanceTypeDetails *nodespec.InstanceType
	sumCPUs             int64
	sumMemory           float64
}

// collectArguments creates arguments based on command line flags and config.
func collectArguments(cmdLineArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		clusterID:         cmdLineArgs[0],
		scheme:            scheme,
		userProvidedToken: flags.Token,
	}
}

func verifyPreconditions(args Arguments, positionalArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)
	err := verifyPreconditions(args, positionalArgs)
	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
}

// fetchNodePools collects all information we would want to display
// on the node pools of a cluster.
func fetchNodePools(args Arguments) ([]*resultRow, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetNodePools(args.clusterID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// sort node pools by ID
	sort.Slice(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// create combined output data structure.
	rows := []*resultRow{}

	for _, np := range response.Payload {
		it, err := awsInfo.GetInstanceTypeDetails(np.NodeSpec.Aws.InstanceType)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		sumCPUs := np.Status.NodesReady * int64(it.CPUCores)
		sumMemory := float64(np.Status.NodesReady) * float64(it.MemorySizeGB)

		rows = append(rows, &resultRow{np, it, sumCPUs, sumMemory})
	}

	return rows, nil

}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)
	nodePools, err := fetchNodePools(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		if clientErr, ok := err.(*clienterror.APIError); ok {
			fmt.Println(color.RedString(clientErr.ErrorMessage))
			if clientErr.ErrorDetails != "" {
				fmt.Println(clientErr.ErrorDetails)
			}
		} else {
			fmt.Println(color.RedString("Error: %s", err.Error()))
		}
		os.Exit(1)
	}

	if len(nodePools) == 0 {
		fmt.Println(color.YellowString("This cluster has no node pools"))
		return
	}

	table := []string{}

	headers := []string{
		color.CyanString("ID"),
		color.CyanString("NAME"),
		color.CyanString("AZ"),
		color.CyanString("INSTANCE TYPE"),
		color.CyanString("NODES MIN/MAX"),
		color.CyanString("NODES DESIRED"),
		color.CyanString("NODES READY"),
		color.CyanString("CPUS"),
		color.CyanString("RAM (GB)"),
	}
	table = append(table, strings.Join(headers, "|"))

	for _, row := range nodePools {
		table = append(table, strings.Join([]string{
			row.nodePool.ID,
			row.nodePool.Name,
			formatting.AvailabilityZonesList(row.nodePool.AvailabilityZones),
			row.nodePool.NodeSpec.Aws.InstanceType,
			strconv.FormatInt(row.nodePool.Scaling.Min, 10) + "/" + strconv.FormatInt(row.nodePool.Scaling.Max, 10),
			strconv.FormatInt(row.nodePool.Status.Nodes, 10),
			formatNodesReady(row.nodePool.Status.Nodes, row.nodePool.Status.NodesReady),
			strconv.FormatInt(row.sumCPUs, 10),
			strconv.FormatFloat(row.sumMemory, 'f', 1, 64),
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
	}

	fmt.Println(columnize.Format(table, colConfig))
}

func formatNodesReady(nodes, nodesReady int64) string {
	if nodes == nodesReady {
		return strconv.FormatInt(nodesReady, 10)
	}

	return color.YellowString(strconv.FormatInt(nodesReady, 10))
}
