// Package nodepools implements the 'list organizations' sub-command.
package nodepools

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command performs the "list organizations" function
	Command = &cobra.Command{
		Use:     "nodepools <cluster-id>",
		Aliases: []string{"nps", "np"},

		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "List node pools",
		Long: `Prints a list of the node pools of a cluster.

The result will be a table of all node pools of a specific cluster with some details.

To see all available details for a cluster, use 'gsctl show nodepool <cluster-id>/<nodepool-id>'.

To list all clusters you have access to, use 'gsctl list clusters'.`,
		PreRun: printValidation,
		Run:    printResult,
	}
)

const activityName = "list-nodepools"

type arguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
	clusterID   string
}

// defaultArgs creates arguments based on command line flags and config.
func defaultArgs(cmdLineArgs []string) arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return arguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
		clusterID:   cmdLineArgs[0],
	}
}

func verifyPreconditions(args arguments, positionalArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := defaultArgs(positionalArgs)
	err := verifyPreconditions(args, positionalArgs)
	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := defaultArgs(positionalArgs)
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

	table := []string{}

	headers := []string{
		color.CyanString("ID"),
		color.CyanString("NAME"),
		color.CyanString("AZ"),
		color.CyanString("INSTANCE TYPE"),
		color.CyanString("NODES MIN/MAX"),
		color.CyanString("NODES DESIRED"),
		color.CyanString("NODES READY"),
	}
	table = append(table, strings.Join(headers, "|"))

	for _, np := range nodePools {
		table = append(table, strings.Join([]string{
			np.ID,
			np.Name,
			formatAvailabilityZones(np.AvailabilityZones),
			np.NodeSpec.Aws.InstanceType,
			"TODO",
			string(np.Status.Nodes),
			string(np.Status.NodesReady),
		}, "|"))
	}

	fmt.Println(columnize.SimpleFormat(table))
}

func fetchNodePools(args arguments) (models.V5GetNodePoolsResponse, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientV2.GetNodePools(args.clusterID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// sort node pools by ID
	sort.Slice(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	return response.Payload, nil

}

// formatAvailabilityZones returns the list of availability zones
// as one string consisting of uppercase letters only, e. g. "A,B,C".
func formatAvailabilityZones(az []string) string {
	shortened := []string{}

	for _, az := range az {
		// last character of each item
		shortened = append(shortened, az[len(az)-1:])
	}

	sort.Strings(shortened)

	return strings.ToUpper(strings.Join(shortened, ","))
}
