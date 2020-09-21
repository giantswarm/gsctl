// Package nodepool implements the "update nodepool" command.
package nodepool

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the cobra command for 'gsctl update nodepool'
	Command = &cobra.Command{
		Use:     "nodepool <cluster-name/cluster-id>/<nodepool-id>",
		Aliases: []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Modify node pool details",
		Long: `Change the name or the scaling settings of a node pool.

Setting the '--nodes-min' and '--nodes-max' flags to different values will enable autoscaling (AWS only).

Examples:

  gsctl update nodepool f01r4/75rh1 --name "General purpose M5"

  gsctl update nodepool f01r4/75rh1 --nodes-min 10 --nodes-max 20

  gsctl update nodepool "Cluster name"/75rh1 --nodes-min 10 --nodes-max 20

`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	arguments Arguments
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
	ClusterNameOrID   string
	Name              string
	NodePoolID        string
	ScalingMax        int64
	ScalingMin        int64
	Provider          string
	UserProvidedToken string
}

func collectArguments(positionalArgs []string) (Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	parts := strings.Split(positionalArgs[0], "/")
	if len(parts) != 2 {
		return Arguments{}, microerror.Mask(errors.NodePoolIDMalformedError)
	}

	var err error
	var info *models.V4InfoResponse
	{
		if flags.Verbose {
			fmt.Println(color.WhiteString("Fetching installation info to validate input"))
		}

		info, err = getInstallationInfo(endpoint, flags.Token)
		if err != nil {
			return Arguments{}, microerror.Mask(err)
		}
	}

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterNameOrID:   strings.TrimSpace(parts[0]),
		Name:              flags.Name,
		NodePoolID:        strings.TrimSpace(parts[1]),
		ScalingMax:        flags.WorkersMax,
		ScalingMin:        flags.WorkersMin,
		Provider:          info.General.Provider,
		UserProvidedToken: flags.Token,
	}, nil
}

// result represents all information we get back from modifying a node pool.
type result struct {
	// nodePool contains all the node pool details as returned from the API.
	NodePool *models.V5GetNodePoolResponse
}

func verifyPreconditions(args Arguments) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	} else if args.ClusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	} else if args.NodePoolID == "" {
		return microerror.Mask(errors.NodePoolIDMissingError)
	}

	if args.ScalingMin == 0 && args.ScalingMax == 0 && args.Name == "" {
		return microerror.Maskf(errors.NoOpError, "Nothing to update.")
	}

	if args.ScalingMin > args.ScalingMax && args.ScalingMax > 0 {
		return microerror.Mask(errors.WorkersMinMaxInvalidError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	var err error
	arguments, err = collectArguments(positionalArgs)

	if err == nil {
		err = verifyPreconditions(arguments)
	}

	if err == nil {
		return
	}

	handleError(err)
	os.Exit(1)
}

func updateNodePool(args Arguments) (*result, error) {
	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.APIEndpoint, args.ClusterNameOrID, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName
	existingNP, err := clientWrapper.GetNodePool(clusterID, args.NodePoolID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	updated := false
	modifyRequestBody := &models.V5ModifyNodePoolRequest{}
	{
		if args.Name != "" && args.Name != existingNP.Payload.Name {
			modifyRequestBody.Name = args.Name
			updated = true
		}
		if args.ScalingMin != 0 && args.ScalingMin != existingNP.Payload.Scaling.Min ||
			args.ScalingMax != 0 && args.ScalingMax != existingNP.Payload.Scaling.Max {
			modifyRequestBody.Scaling = &models.V5ModifyNodePoolRequestScaling{
				Min: args.ScalingMin,
				Max: args.ScalingMax,
			}
			updated = true
		}
	}
	if !updated {
		return nil, microerror.Maskf(errors.NoOpError, "Nothing to update.")
	}

	auxParams.ActivityName = activityName
	response, err := clientWrapper.ModifyNodePool(clusterID, args.NodePoolID, modifyRequestBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &result{
		NodePool: response.Payload,
	}

	return r, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	r, err := updateNodePool(arguments)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}

	fmt.Println(color.GreenString("Node pool '%s' (ID '%s') in cluster '%s' has been modified.", r.NodePool.Name, r.NodePool.ID, arguments.ClusterNameOrID))
}

func getInstallationInfo(endpoint, userProvidedToken string) (*models.V4InfoResponse, error) {
	clientWrapper, err := client.NewWithConfig(endpoint, userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	installationInfo, err := clientWrapper.GetInfo(nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return installationInfo.Payload, nil
}

func handleError(err error) {
	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	case errors.IsNodePoolIDMalformedError(err):
		headline = "Bad format for Cluster name/ID or Node Pool ID argument"
		subtext = "Please provide cluster name/ID and node pool ID separated by a slash. See --help for examples."
	case errors.IsNoOpError(err):
		headline = microerror.Pretty(err, false)
	default:
		headline = err.Error()
	}

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
}
