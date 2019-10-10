// Package cluster implements the 'scale cluster' command.
package cluster

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/limits"
)

var (
	// Command performs the "delete cluster" function
	Command = &cobra.Command{
		Use:   "cluster",
		Short: "Scale cluster",
		Long: `Increase or reduce the number of worker nodes in a cluster.

Caution:

When reducing the number of nodes, the selection of the node(s) to be removed
is non-deterministic. Workloads on the worker nodes to be removed will be
terminated, data stored on the worker nodes will be lost. Make sure to remove
only as many nodes as your deployment architecture can handle in a resilient
way.

Examples:

  gsctl scale cluster c7t2o --workers-min 12 --workers-max 16

  gsctl scale cluster c7t2o --workers-min 3 --workers-max 3

  gsctl scale cluster c7t2o --num-workers 3
`,

		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	//Flag names.
	cmdWorkersMaxName = "workers-max"
	cmdWorkersMinName = "workers-min"
	cmdWorkersNumName = "num-workers"
)

const (
	scaleClusterActivityName = "scale-cluster"
)

// Arguments contains all arguments that influence the business function.
type Arguments struct {
	apiEndpoint         string
	authToken           string
	clusterID           string
	numWorkersDesired   int
	oppressConfirmation bool
	scheme              string
	userProvidedToken   string
	verbose             bool
	workersMax          int64
	workersMaxSet       bool
	workersMin          int64
	workersMinSet       bool
	workers             int
	workersSet          bool
}

// Result is the resulting data we get from our business function.
type Result struct {
	numWorkersBefore int
	scalingMinBefore int
	scalingMinAfter  int
	scalingMaxBefore int
	scalingMaxAfter  int
}

func init() {
	Command.Flags().BoolVarP(&flags.Force, "force", "", false, "If set, no confirmation is required.")
	Command.Flags().Int64VarP(&flags.WorkersMax, cmdWorkersMaxName, "", 0, "Maximum number of worker nodes to have after scaling.")
	Command.Flags().Int64VarP(&flags.WorkersMin, cmdWorkersMinName, "", 0, "Minimum number of worker nodes to have after scaling.")
	Command.Flags().IntVarP(&flags.NumWorkers, cmdWorkersNumName, "w", 0, "Shorthand to set --workers-min and --workers-max to the same value.")
}

// getConfirmation asks the user for confirmation for scaling actions.
func getConfirmation(args Arguments, maxBefore int, minBefore int, currentWorkers int) error {
	if int64(currentWorkers) > args.workersMax && args.workersMax == args.workersMin {
		confirmed := confirm.Ask(fmt.Sprintf("The cluster currently has %d worker nodes running.\nDo you want to pin the number of worker nodes to %d?", currentWorkers, args.workersMin))
		if !confirmed {
			return microerror.Mask(errors.CommandAbortedError)
		}
	}
	if int64(currentWorkers) > args.workersMax && args.workersMax != args.workersMin {
		confirmed := confirm.Ask(fmt.Sprintf("The cluster currently has %d worker nodes running.\nDo you want to change the limits to be min=%d, max=%d?", currentWorkers, args.workersMin, args.workersMax))
		if !confirmed {
			return microerror.Mask(errors.CommandAbortedError)
		}
	}

	return nil
}

func collectArguments(cmd *cobra.Command, positionalArgs []string) (Arguments, error) {
	if len(positionalArgs) == 0 {
		return Arguments{}, microerror.Mask(errors.ClusterIDMissingError)
	}

	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	args := Arguments{
		apiEndpoint:         endpoint,
		authToken:           token,
		clusterID:           positionalArgs[0],
		oppressConfirmation: flags.Force,
		scheme:              scheme,
		userProvidedToken:   flags.Token,
		verbose:             flags.Verbose,
		workersMax:          flags.WorkersMax,
		workersMin:          flags.WorkersMin,
		workers:             flags.NumWorkers,
		workersMaxSet:       cmd.Flags().Changed(cmdWorkersMaxName),
		workersMinSet:       cmd.Flags().Changed(cmdWorkersMinName),
		workersSet:          cmd.Flags().Changed(cmdWorkersNumName),
	}

	if args.workers > 0 {
		args.workersMin = int64(args.workers)
		args.workersMax = int64(args.workers)
	}

	return args, nil
}

func verifyPreconditions(args Arguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.clusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	if args.workersSet && (args.workersMinSet || args.workersMaxSet) {
		return microerror.Mask(errors.ConflictingWorkerFlagsUsedError)
	}
	if args.workersMax > 0 && args.workersMax < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.CannotScaleBelowMinimumWorkersError)
	}
	if args.workersMin > 0 && args.workersMin < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.workers != 0 && args.workers < limits.MinimumNumWorkers {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if !args.workersSet && !args.workersMinSet && !args.workersMaxSet {
		return microerror.Maskf(errors.RequiredFlagMissingError, "--%s or --%s/--%s", cmdWorkersNumName, cmdWorkersMinName, cmdWorkersMaxName)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args, err := collectArguments(cmd, positionalArgs)
	if err != nil {
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	err = verifyPreconditions(args)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	case errors.IsConflictingWorkerFlagsUsed(err):
		headline = "Conflicting flags used"
		subtext = fmt.Sprintf("When specifying --%s, neither --%s nor --%s must be used.", cmdWorkersNumName, cmdWorkersMaxName, cmdWorkersMinName)
	case errors.IsWorkersMinMaxInvalid(err):
		headline = "Number of worker nodes invalid"
		subtext = fmt.Sprintf("Node count flag --%s must not be higher than --%s.", cmdWorkersMinName, cmdWorkersMaxName)
	case errors.IsCannotScaleBelowMinimumWorkersError(err):
		headline = "Not enough worker nodes specified"
		subtext = fmt.Sprintf("You'll need at least %v worker nodes for a useful cluster.", limits.MinimumNumWorkers)
	case errors.IsRequiredFlagMissingError(err):
		headline = "Missing flag: " + err.Error()
		subtext = "Please use --help to see details regarding the command's usage."
	default:
		headline = err.Error()
	}

	// handle non-common errors
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

// scaleCluster is the actual function submitting the API call and handling the response.
func scaleCluster(args Arguments) (*Result, error) {
	clientWrapper, err := client.NewWithConfig(flags.APIEndpoint, flags.Token)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = scaleClusterActivityName

	// Make sure we have provider info in the current endpoint
	if config.Config.Provider == "" {
		if args.verbose {
			fmt.Println(color.WhiteString("Fetching provider information"))
		}

		info, err := clientWrapper.GetInfo(auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		config.Config.SetProvider(info.Payload.General.Provider)
	}

	if args.verbose {
		fmt.Println(color.WhiteString("Fetching v4 cluster details"))
	}
	clusterDetails, err := clientWrapper.GetClusterV4(args.clusterID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	scalingResult := &Result{
		numWorkersBefore: int(len(clusterDetails.Payload.Workers)),
		scalingMaxBefore: int(clusterDetails.Payload.Scaling.Max),
		scalingMinBefore: int(clusterDetails.Payload.Scaling.Min),
	}

	var statusWorkers int

	if args.verbose {
		fmt.Println(color.WhiteString("Fetching v4 cluster status"))
	}

	status, err := clientWrapper.GetClusterStatus(args.clusterID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(status.Cluster.Nodes) >= 1 {
		// Count all nodes as workers which are not explicitly marked as master.
		for _, node := range status.Cluster.Nodes {
			val, ok := node.Labels["role"]
			if ok && val == "master" {
				// don't count this
			} else {
				statusWorkers++
			}
		}
	}

	// Ask for confirmation for the scaling action.
	if !flags.Force {
		// get confirmation and handle result
		err = getConfirmation(args, scalingResult.scalingMaxBefore, scalingResult.scalingMinBefore, statusWorkers)
		if err != nil {
			fmt.Println(color.GreenString("Scaling cancelled"))
			os.Exit(0)
		}
	}

	// Preparing API call.
	reqBody := &models.V4ModifyClusterRequest{
		Scaling: &models.V4ModifyClusterRequestScaling{
			Max: args.workersMax,
			Min: args.workersMin,
		},
	}

	// perform API call
	if args.verbose {
		fmt.Println(color.WhiteString("Sending API request to modify cluster"))
	}
	_, err = clientWrapper.ModifyCluster(args.clusterID, reqBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	scalingResult.scalingMinAfter = int(args.workersMin)
	scalingResult.scalingMaxAfter = int(args.workersMax)

	return scalingResult, nil
}

// printResult invokes the actual cluster scaling and prints the result and/or errors.
func printResult(cmd *cobra.Command, commandLineArgs []string) {
	args, err := collectArguments(cmd, commandLineArgs)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	// Actually make the scaling request to the API.
	result, err := scaleCluster(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		headline := ""
		subtext := ""

		switch {
		case errors.IsCannotScaleBelowMinimumWorkersError(err):
			headline = "Desired worker node count is too low."
			subtext = fmt.Sprintf("Please set the -w|--%s or --%s flag to a value greater than 0.", cmdWorkersNumName, cmdWorkersMinName)
		case errors.IsDesiredEqualsCurrentStateError(err):
			headline = "Desired worker node count equals the current one."
			subtext = "No worker nodes have been added or removed."
		case errors.IsCouldNotScaleClusterError(err):
			headline = "The cluster could not be scaled."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"
		case errors.IsCommandAbortedError(err):
			headline = "Cancelled"
			subtext = "Scaling settings of this cluster stay as they are."
		default:
			headline = err.Error()
		}

		// Print error output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	fmt.Println(color.GreenString("The cluster is being scaled"))
	fmt.Printf("The cluster limits have been changed from min=%d and max=%d to min=%d and max=%d workers.\n", result.scalingMinBefore, result.scalingMaxBefore, result.scalingMinAfter, result.scalingMaxAfter)
}
