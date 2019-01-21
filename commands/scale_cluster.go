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

		// PreRun checks a few general things, like authentication.
		PreRun: scaleClusterValidationOutput,

		// Run calls the business function and prints results and errors.
		Run: scaleClusterRunOutput,
	}
)

const (
	scaleClusterActivityName = "scale-cluster"
)

type ClusterStatus struct {
	Cluster *v1alpha1.StatusCluster `json:"cluster,omitempty"`
}

type scaleClusterArguments struct {
	clusterID           string
	numWorkersDesired   int
	oppressConfirmation bool
	verbose             bool
	apiEndpoint         string
	authToken           string
	scheme              string
	workerNumCPUs       int
	workerMemorySizeGB  float32
	workerStorageSizeGB float32
	workersMax          int64
	workersMin          int64
}

func defaultScaleClusterArguments() scaleClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return scaleClusterArguments{
		apiEndpoint:         endpoint,
		authToken:           token,
		scheme:              scheme,
		clusterID:           cmdClusterID,
		numWorkersDesired:   cmdNumWorkers,
		workerNumCPUs:       cmdWorkerNumCPUs,
		workerMemorySizeGB:  cmdWorkerMemorySizeGB,
		workerStorageSizeGB: cmdWorkerStorageSizeGB,
		workersMax:          cmdWorkersMax,
		workersMin:          cmdWorkersMin,
		oppressConfirmation: cmdForce,
		verbose:             cmdVerbose,
	}
}

type scaleClusterResults struct {
	// min number of workers according to our info, just before the PATCH call.
	workersMinBefore int64
	// max number of workers according to our info, just before the PATCH call.
	workersMaxBefore int64
	// min number of workers as of the PATCH call response.
	workersMinAfter int64
	// max number of workers as of the PATCH call response.
	workersMaxAfter int64
}

func init() {
	ScaleClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no confirmation is required when reducing the number of workers.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMin, "workers-min", "", 0, "Minimum number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().Int64VarP(&cmdWorkersMax, "workers-max", "", 0, "Maximum number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().IntVarP(&cmdNumWorkers, "num-workers", "w", 0, "Number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, "num-cpus", "", 0, "Number of CPU cores per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, "memory-gb", "", 0, "RAM per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per added worker node.")

	CreateClusterCommand.Flags().MarkDeprecated("num-workers", "please use --workers-min and --workers-max to specify the node count to use")

	ScaleCommand.AddCommand(ScaleClusterCommand)
}

// scaleClusterValidationOutput calls a pre-check function. In case anything is missing,
// displays the error and exits with code 1.
func scaleClusterValidationOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultScaleClusterArguments()

	headline := ""
	subtext := ""

	err := validateScaleCluster(args, cmdLineArgs)
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
}

// validatyScaleCluster does a few general checks and returns an error in case something is missing.
func validateScaleCluster(args scaleClusterArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(clusterIDMissingError)
	}

	if args.numWorkersDesired > 0 && (args.workersMax > 0 || args.workersMin > 0) {
		return microerror.Mask(conflictingWorkerFlagsUsedError)
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

// scaleClusterRunOutput invokes the actual cluster scaling and prints the result and/or errors.
func scaleClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultScaleClusterArguments()
	args.clusterID = cmdLineArgs[0]

	_, err := scaleCluster(args)
	if err != nil {
		handleCommonErrors(err)

		var headline = ""
		var subtext = ""

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

}

// scaleCluster is the actual function submitting the API call and handling the response.
func scaleCluster(args scaleClusterArguments) (scaleClusterResults, error) {
	results := scaleClusterResults{}

	clusterDetails, err := getClusterDetails(args.clusterID, scaleClusterActivityName)
	if err != nil {
		return results, microerror.Mask(err)
	}
	status, err := getDesiredCapacity(args.clusterID, scaleClusterActivityName)
	if err != nil {
		return results, microerror.Mask(err)
	}
	results.workersMaxBefore = clusterDetails.Scaling.Max
	results.workersMinBefore = clusterDetails.Scaling.Min

	if results.workersMaxBefore == args.workersMax && results.workersMinAfter == args.workersMin {
		return results, microerror.Mask(desiredEqualsCurrentStateError)
	}

	// confirmation in case of scaling down
	if !args.oppressConfirmation && int64(status.Scaling.DesiredCapacity) > args.workersMax {
		confirmed := askForConfirmation(fmt.Sprintf("Do you really want to reduce the worker nodes for cluster '%s' to %d?",
			args.clusterID, args.numWorkersDesired))
		if !confirmed {
			return results, microerror.Mask(commandAbortedError)
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
		fmt.Println("Sending API request to modify cluster")
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = scaleClusterActivityName

	_, err = ClientV2.ModifyCluster(args.clusterID, reqBody, auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return results, microerror.Mask(accessForbiddenError)
			} else if clientErr.HTTPStatusCode == http.StatusNotFound {
				return results, microerror.Mask(clusterNotFoundError)
			}
		}

		return results, microerror.Mask(err)
	}

	return results, nil
}

// getClusterStatus returns status for one cluster.
func getDesiredCapacity(clusterID, activityName string) (*v1alpha1.StatusCluster, error) {
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
