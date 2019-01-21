package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
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

  gsctl scale cluster c7t2o -w 12

  gsctl scale cluster c7t2o -w 12 --num-cpus 4
`,

		// Run calls the business function and prints results and errors.
		Run: scaleClusterRunOutput,
	}

	//Flag names.
	cmdWorkersMinName            = "workers-min"
	cmdWorkersMaxName            = "workers-max"
	cmdWorkerNumCPUsName         = "num-cpus"
	cmdKVMWorkerNumCPUsName      = "kvm-num-cpus"
	cmdWorkerMemorySizeGBName    = "memory-gb"
	cmdKVMWorkerMemorySizeGBName = "kvm-memory-gb"
)

const (
	scaleClusterActivityName = "scale-cluster"
)

type scaleClusterArguments struct {
	clusterID             string
	numWorkersDesired     int
	oppressConfirmation   bool
	verbose               bool
	apiEndpoint           string
	authToken             string
	scheme                string
	workerNumCPUs         int
	workerMemorySizeGB    float32
	workerKVMNumCPUs      int
	workerKVMMemorySizeGB float32
	workerStorageSizeGB   float32
	workersMax            int64
	workersMin            int64
}

func init() {
	ScaleClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no confirmation is required when reducing the number of workers.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMin, cmdWorkersMinName, "", 0, "Minimum number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMax, cmdWorkersMaxName, "", 0, "Maximum number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdKVMWorkerMemorySizeGB, cmdKVMWorkerMemorySizeGBName, "", 0, "RAM per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per added worker node.")
	ScaleClusterCommand.Flags().IntVarP(&cmdKVMWorkerNumCPUs, cmdKVMWorkerNumCPUsName, "", 0, "Number of CPU cores per added worker node.")

	ScaleClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, cmdWorkerNumCPUsName, "", 0, "Number of CPU cores per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, cmdWorkerMemorySizeGBName, "", 0, "RAM per added worker node.")
	ScaleClusterCommand.Flags().IntVarP(&cmdNumWorkers, "num-workers", "w", 0, "Number of worker nodes to have after scaling.")

	ScaleClusterCommand.Flags().MarkDeprecated("num-workers", "Please use --workers-min and --workers-max to specify the node count to use.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkerNumCPUsName, "Please use --kvm-num-cpus to specify the cpu number for kvm to use.")
	ScaleClusterCommand.Flags().MarkDeprecated(cmdWorkerMemorySizeGBName, "Please use --kvm-num-memory-gb to specify the amount of memory for kvm to use.")

	ScaleCommand.AddCommand(ScaleClusterCommand)
}

func defaultScaleClusterArguments(cmd *cobra.Command, clusterId string, maxBefore int64, minBefore int64) scaleClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	scaleArgs := scaleClusterArguments{
		apiEndpoint:           endpoint,
		authToken:             token,
		scheme:                scheme,
		clusterID:             clusterId,
		numWorkersDesired:     cmdNumWorkers,
		workerKVMNumCPUs:      cmdKVMWorkerNumCPUs,
		workerKVMMemorySizeGB: cmdKVMWorkerMemorySizeGB,
		workerNumCPUs:         cmdWorkerNumCPUs,
		workerMemorySizeGB:    cmdWorkerMemorySizeGB,
		workerStorageSizeGB:   cmdWorkerStorageSizeGB,
		workersMax:            cmdWorkersMax,
		workersMin:            cmdWorkersMin,
		oppressConfirmation:   cmdForce,
		verbose:               cmdVerbose,
	}

	if !cmd.Flags().Changed(cmdWorkersMinName) {
		scaleArgs.workersMin = minBefore
	}
	if !cmd.Flags().Changed(cmdWorkersMaxName) {
		scaleArgs.workersMin = maxBefore
	}
	if !cmd.Flags().Changed(cmdWorkersMaxName) && !cmd.Flags().Changed(cmdWorkersMinName) && cmd.Flags().Changed("num-workers") {
		scaleArgs.workersMax = int64(scaleArgs.numWorkersDesired)
		scaleArgs.workersMin = int64(scaleArgs.numWorkersDesired)
	}
	if cmd.Flags().Changed(cmdWorkerMemorySizeGBName) && !cmd.Flags().Changed(cmdKVMWorkerMemorySizeGBName) {
		scaleArgs.workerKVMMemorySizeGB = cmdWorkerMemorySizeGB
	}
	if cmd.Flags().Changed(cmdWorkerNumCPUsName) && !cmd.Flags().Changed(cmdKVMWorkerNumCPUsName) {
		scaleArgs.workerKVMNumCPUs = cmdWorkerNumCPUs
	}

	return scaleArgs
}

// validatyScaleCluster does a few general checks and returns an error in case something is missing.
func validateScaleCluster(args scaleClusterArguments, cmdLineArgs []string, maxBefore int64, minBefore int64, desiredCapacity int) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}

	if maxBefore == args.workersMax && minBefore == args.workersMin {
		return microerror.Mask(desiredEqualsCurrentStateError)
	}

	// flag conflicts.
	if args.numWorkersDesired > 0 && (args.workersMax > 0 || args.workersMin > 0) {
		return microerror.Mask(conflictingWorkerFlagsUsedError)
	}
	if args.workerNumCPUs > 0 && args.workerKVMNumCPUs > 0 && args.workerNumCPUs != args.workerKVMNumCPUs {
		return microerror.Mask(conflictingFlagsError)
	}
	if args.workerMemorySizeGB > 0 && args.workerKVMMemorySizeGB > 0 && args.workerMemorySizeGB != args.workerKVMMemorySizeGB {
		return microerror.Mask(conflictingFlagsError)
	}

	if args.numWorkersDesired > 0 && args.numWorkersDesired < minimumNumWorkers {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMax > 0 && args.workersMax < int64(minimumNumWorkers) {
		return microerror.Mask(cannotScaleBelowMinimumWorkersError)
	}
	if args.workersMin > 0 && args.workersMin < int64(minimumNumWorkers) {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMin > 0 && args.workersMax > 0 && args.workersMin > args.workersMax {
		return microerror.Mask(workersMinMaxInvalidError)
	}

	return nil
}

func confirmScaleCluster(args scaleClusterArguments, maxBefore int64, minBefore int64, desiredCapacity int) error {
	fmt.Printf("Do you really want to reduce the worker nodes for cluster '%s' to %d?",
		args.clusterID, args.numWorkersDesired)
	// confirmation in case of scaling down
	if !args.oppressConfirmation && int64(desiredCapacity) > args.workersMax {
		confirmed := askForConfirmation(fmt.Sprintf("Do you really want to reduce the worker nodes for cluster '%s' to %d?",
			args.clusterID, args.numWorkersDesired))
		if !confirmed {
			return microerror.Mask(commandAbortedError)
		}
	}

	return nil

}

// scaleClusterRunOutput invokes the actual cluster scaling and prints the result and/or errors.
func scaleClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	if len(cmdLineArgs) == 0 {
		handleCommonErrors(clusterIDMissingError)
	}
	clusterDetails, err := getClusterDetails(cmdLineArgs[0], scaleClusterActivityName)
	if err != nil {
		fmt.Println(color.RedString("Error getting cluster details!"))
		handleCommonErrors(err)
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}
	status, err := getClusterStatus(cmdLineArgs[0], scaleClusterActivityName)
	if err != nil {
		fmt.Println(color.RedString("Error getting cluster status!"))
		handleCommonErrors(err)
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	args := defaultScaleClusterArguments(cmd, cmdLineArgs[0], clusterDetails.Scaling.Max, clusterDetails.Scaling.Min)

	headline := ""
	subtext := ""

	err = validateScaleCluster(args, cmdLineArgs, clusterDetails.Scaling.Max, clusterDetails.Scaling.Min, status.Scaling.DesiredCapacity)
	if err != nil {
		handleCommonErrors(err)

		switch {
		case IsConflictingWorkerFlagsUsed(err):
			headline = "Conflicting flags used"
			subtext = "When specifying --num-workers, neither --workers-max nor --workers-min must be used."
		case IsWorkersMinMaxInvalid(err):
			headline = "Number of worker nodes invalid"
			subtext = "Node count flag --workers-min must not be higher than --workers-max."
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

	err = confirmScaleCluster(args, clusterDetails.Scaling.Max, clusterDetails.Scaling.Min, status.Scaling.DesiredCapacity)
	if err != nil {
		handleCommonErrors(err)
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	err = scaleCluster(args)
	if err != nil {
		handleCommonErrors(err)

		switch {
		case IsCommandAbortedError(err):
			headline = "Scaling cancelled."
		case IsCannotScaleBelowMinimumWorkersError(err):
			headline = "Desired worker node count is too low."
			subtext = "Please set the -w|--num-workers flag to a value greater than 0."
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
	fmt.Printf("The cluster limits have been changed from min = %d and max = %d to min = %d and max = %d workers.\n", clusterDetails.Scaling.Min, clusterDetails.Scaling.Max, args.workersMin, args.workersMax)

}

// scaleCluster is the actual function submitting the API call and handling the response.
func scaleCluster(args scaleClusterArguments) error {
	// Preparing API call.
	reqBody := &models.V4ModifyClusterRequest{
		Scaling: &models.V4ModifyClusterRequestScaling{
			Max: args.workersMax,
			Min: args.workersMin,
		},
		Workers: []*models.V4ModifyClusterRequestWorkersItems{},
	}

	// perform API call
	if args.verbose {
		fmt.Println("Sending API request to modify cluster")
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = scaleClusterActivityName

	_, err := ClientV2.ModifyCluster(args.clusterID, reqBody, auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return microerror.Mask(accessForbiddenError)
			} else if clientErr.HTTPStatusCode == http.StatusNotFound {
				return microerror.Mask(clusterNotFoundError)
			}
		}

		return microerror.Mask(err)
	}

	return nil
}

// getClusterStatus returns status for one cluster.
func getClusterStatus(clusterID, activityName string) (*v1alpha1.StatusCluster, error) {
	// perform API call
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := ClientV2.GetClusterStatus(clusterID, auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			switch clientErr.HTTPStatusCode {
			case http.StatusForbidden:
				return nil, microerror.Mask(accessForbiddenError)
			case http.StatusUnauthorized:
				return nil, microerror.Mask(notAuthorizedError)
			case http.StatusNotFound:
				return nil, microerror.Mask(clusterNotFoundError)
			case http.StatusInternalServerError:
				return nil, microerror.Mask(internalServerError)
			}
		}

		return nil, microerror.Mask(err)
	}

	// We have to marshal and unmarshal here because the generated client gives us
	// a map[string]interface and we want to unmarshal it into the actual
	// apiextensions type.
	m, err := json.Marshal(response.Payload)
	if err != nil {
		return nil, err
	}

	var status v1alpha1.StatusCluster

	err = json.Unmarshal(m, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
