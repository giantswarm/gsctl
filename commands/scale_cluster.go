package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// ScaleClusterCommand performs the "delete cluster" function
	ScaleClusterCommand = &cobra.Command{
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
`,

		// Run calls the business function and prints results and errors.
		Run: scaleClusterRunOutput,
	}

	//Flag names.
	cmdWorkersMaxName = "workers-max"
	cmdWorkersMinName = "workers-min"
	cmdWorkersNumName = "num-workers"

	cmdWorkerMemorySizeGBName  = "memory-gb"
	cmdWorkerNumCPUsName       = "num-cpus"
	cmdWorkerStorageSizeGBName = "storage-gb"
)

const (
	scaleClusterActivityName = "scale-cluster"
)

type scaleClusterArguments struct {
	apiEndpoint         string
	authToken           string
	clusterID           string
	numWorkersDesired   int
	oppressConfirmation bool
	scheme              string
	verbose             bool
	workersMax          int64
	workersMin          int64
}

func init() {
	ScaleClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no confirmation is required.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMax, cmdWorkersMaxName, "", 0, "Maximum number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMin, cmdWorkersMinName, "", 0, "Minimum number of worker nodes to have after scaling.")

	// deprecated
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, cmdWorkerStorageSizeGBName, "", 0, "Local storage size per added worker node.")
	ScaleClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, cmdWorkerNumCPUsName, "", 0, "Number of CPU cores per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, cmdWorkerMemorySizeGBName, "", 0, "RAM per added worker node.")
	ScaleClusterCommand.Flags().IntVarP(&cmdNumWorkers, cmdWorkersNumName, "w", 0, "Number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkersNumName, "Please use --workers-min and --workers-max to specify the node count to use.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkerMemorySizeGBName, "Changing the amount of Memory is no longer supported while scaling.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkerNumCPUsName, "Changing the number of CPUs is no longer supported while scaling.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkerStorageSizeGBName, "Changing the amount of Storage is no longer supported while scaling.")

	ScaleCommand.AddCommand(ScaleClusterCommand)
}

// confirmScaleCluster asks the user for confirmation for scaling actions.
func confirmScaleCluster(args scaleClusterArguments, maxBefore int64, minBefore int64, desiredCapacity int64) error {
	if desiredCapacity > args.workersMax && args.workersMax == args.workersMin {
		confirmed := askForConfirmation(fmt.Sprintf("The cluster currently has %d worker nodes running.\nDo you want to pin the number of worker nodes to %d?", desiredCapacity, args.workersMin))
		if !confirmed {
			return microerror.Mask(commandAbortedError)
		}
	}
	if desiredCapacity > args.workersMax && args.workersMax != args.workersMin {
		confirmed := askForConfirmation(fmt.Sprintf("The cluster currently has %d worker nodes running.\nDo you want to change the limits to be min=%d, max=%d?", desiredCapacity, args.workersMin, args.workersMax))
		if !confirmed {
			return microerror.Mask(commandAbortedError)
		}
	}

	return nil

}

// defaultScaleClusterArguments defaults arguments supplied by the users.
func defaultScaleClusterArguments(cmd *cobra.Command, clusterId string, maxBefore int64, minBefore int64) scaleClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	scaleArgs := scaleClusterArguments{
		apiEndpoint:         endpoint,
		authToken:           token,
		clusterID:           clusterId,
		numWorkersDesired:   cmdNumWorkers,
		oppressConfirmation: cmdForce,
		scheme:              scheme,
		verbose:             cmdVerbose,
		workersMax:          cmdWorkersMax,
		workersMin:          cmdWorkersMin,
	}

	if !cmd.Flags().Changed(cmdWorkersMinName) {
		scaleArgs.workersMin = minBefore
	}
	if !cmd.Flags().Changed(cmdWorkersMaxName) {
		scaleArgs.workersMax = maxBefore
	}
	if !cmd.Flags().Changed(cmdWorkersMaxName) && !cmd.Flags().Changed(cmdWorkersMinName) && cmd.Flags().Changed(cmdWorkersNumName) {
		scaleArgs.workersMax = int64(scaleArgs.numWorkersDesired)
		scaleArgs.workersMin = int64(scaleArgs.numWorkersDesired)
	}

	return scaleArgs
}

func isAutoscalingEnabled(version string) (bool, error) {
	{
		n, err := util.CompareVersions(version, "6.3.0")
		if err != nil {
			return false, err
		}
		if n == 0 || n == 1 {
			return true, nil
		}
	}
	return false, nil
}

// getClusterStatus returns the status for one cluster.
func getClusterStatus(clusterID, activityName string) (*client.ClusterStatus, error) {
	// perform API call
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	status, err := ClientV2.GetClusterStatus(clusterID, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return status, nil
}

// scaleCluster is the actual function submitting the API call and handling the response.
func scaleCluster(args scaleClusterArguments) (*models.V4ClusterDetailsResponse, error) {
	// Preparing API call.
	reqBody := &models.V4ModifyClusterRequest{
		Scaling: &models.V4ModifyClusterRequestScaling{
			Max: args.workersMax,
			Min: args.workersMin,
		},
	}

	// perform API call
	if args.verbose {
		fmt.Println("Sending API request to modify cluster")
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = scaleClusterActivityName

	response, err := ClientV2.ModifyCluster(args.clusterID, reqBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}

// scaleClusterRunOutput invokes the actual cluster scaling and prints the result and/or errors.
func scaleClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	if len(cmdLineArgs) == 0 {
		handleCommonErrors(clusterIDMissingError)
	}
	// Get all necessary information for defaulting and confirmation.
	clusterDetails, err := getClusterDetails(cmdLineArgs[0], scaleClusterActivityName)
	if err != nil {
		fmt.Println(color.RedString("Error getting cluster details!"))
		handleCommonErrors(err)
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	autoScalingEnabled, err := isAutoscalingEnabled(clusterDetails.ReleaseVersion)
	if err != nil {
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	var maxBefore int64
	var minBefore int64
	var desiredCapacity int64

	if autoScalingEnabled {
		// We only need the status if autoscaling is enabled, because we are only
		// interested in the DesiredCapacity.
		status, err := getClusterStatus(cmdLineArgs[0], scaleClusterActivityName)
		if err != nil {
			fmt.Println(color.RedString("Error getting cluster status!"))
			handleCommonErrors(err)
			fmt.Println(color.RedString(err.Error()))
			os.Exit(1)
		}
		maxBefore = clusterDetails.Scaling.Max
		minBefore = clusterDetails.Scaling.Min
		desiredCapacity = int64(status.Cluster.Scaling.DesiredCapacity)

	} else {
		// Default to the length of the workers array. We don't know how old the
		// cluster is and the workers array should be a reliable source of truth for
		// older clusters.
		maxBefore = int64(len(clusterDetails.Workers))
		minBefore = int64(len(clusterDetails.Workers))
		desiredCapacity = int64(len(clusterDetails.Workers))
	}

	// Default all necessary information from flags.
	args := defaultScaleClusterArguments(cmd, cmdLineArgs[0], maxBefore, minBefore)

	headline := ""
	subtext := ""

	// Validate the input for obvious errors.
	err = validateScaleCluster(args, cmdLineArgs, maxBefore, minBefore, desiredCapacity)
	if err != nil {
		handleCommonErrors(err)

		switch {
		case IsConflictingWorkerFlagsUsed(err):
			headline = "Conflicting flags used"
			subtext = fmt.Sprintf("When specifying --%s, neither --%s nor --%s must be used.", cmdWorkersNumName, cmdWorkersMaxName, cmdWorkersMinName)
		case IsWorkersMinMaxInvalid(err):
			headline = "Number of worker nodes invalid"
			subtext = fmt.Sprintf("Node count flag --%s must not be higher than --%s.", cmdWorkersMinName, cmdWorkersMaxName)
		case IsCannotScaleBelowMinimumWorkersError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("You'll need at least %v worker nodes for a useful cluster.", minimumNumWorkers)
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

	// Ask for confirmation for the scaling action.
	if !cmdForce {
		err = confirmScaleCluster(args, maxBefore, minBefore, desiredCapacity)
		if err != nil {
			handleCommonErrors(err)
			fmt.Println(color.RedString(err.Error()))
			os.Exit(1)
		}
	}

	// Actually make the scaling request to the API.
	details, err := scaleCluster(args)
	if err != nil {
		handleCommonErrors(err)

		switch {
		case IsCommandAbortedError(err):
			headline = "Scaling cancelled."
		case IsCannotScaleBelowMinimumWorkersError(err):
			headline = "Desired worker node count is too low."
			subtext = fmt.Sprintf("Please set the -w|--%s or --%s flag to a value greater than 0.", cmdWorkersNumName, cmdWorkersMinName)
		case IsDesiredEqualsCurrentStateError(err):
			headline = "Desired worker node count equals the current one."
			subtext = "No worker nodes have been added or removed."
		case IsCouldNotScaleClusterError(err):
			headline = "The cluster could not be scaled."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"
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
	fmt.Printf("The cluster limits have been changed from min = %d and max = %d to min = %d and max = %d workers.\n", clusterDetails.Scaling.Min, clusterDetails.Scaling.Max, details.Scaling.Min, details.Scaling.Max)

}

// validatyScaleCluster does a few general checks and returns an error in case something is missing.
func validateScaleCluster(args scaleClusterArguments, cmdLineArgs []string, maxBefore int64, minBefore int64, desiredCapacity int64) error {
	desiredWorkersExists := (args.numWorkersDesired > 0)
	scalingParameterIsPresent := (args.workersMax > 0 || args.workersMin > 0)
	desiredWorkersDifferFromMaxNumOfWorkers := (int64(args.numWorkersDesired) != args.workersMax)
	desiredWorkersDifferFromMinNumOfWorkers := (int64(args.numWorkersDesired) != args.workersMin)
	desiredWorkersAreNotAtScalingLimits := (desiredWorkersDifferFromMaxNumOfWorkers || desiredWorkersDifferFromMinNumOfWorkers)

	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}

	if maxBefore == args.workersMax && minBefore == args.workersMin {
		return microerror.Mask(desiredEqualsCurrentStateError)
	}

	// flag conflicts.
	if desiredWorkersExists && scalingParameterIsPresent && desiredWorkersAreNotAtScalingLimits {
		return microerror.Mask(conflictingWorkerFlagsUsedError)
	}

	if desiredWorkersExists && args.numWorkersDesired < minimumNumWorkers {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMax > 0 && args.workersMax < int64(minimumNumWorkers) {
		return microerror.Mask(cannotScaleBelowMinimumWorkersError)
	}
	if args.workersMin > 0 && args.workersMin < int64(minimumNumWorkers) {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if scalingParameterIsPresent && args.workersMin > args.workersMax {
		return microerror.Mask(workersMinMaxInvalidError)
	}

	return nil
}
