package commands

import (
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/fatih/color"
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
		PreRun: scaleClusterPreRunOutput,

		// Run calls the business function and prints results and errors.
		Run: scaleClusterRunOutput,
	}
)

const (
	scaleClusterActivityName = "scale-cluster"
)

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
		oppressConfirmation: cmdForce,
		verbose:             cmdVerbose,
	}
}

type scaleClusterResults struct {
	// number of workers according to our info, just before the PATCH call.
	numWorkersBefore int
	// number of workers to add, just before the PATCH call. Might be negative.
	numWorkersToAdd int
	// number of workers as of the PATCH call response.
	numWorkersAfter int
}

func init() {
	ScaleClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no confirmation is required when reducing the number of workers.")
	ScaleClusterCommand.Flags().IntVarP(&cmdNumWorkers, "num-workers", "w", 0, "Number of worker nodes to have after scaling.")
	ScaleClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, "num-cpus", "", 0, "Number of CPU cores per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, "memory-gb", "", 0, "RAM per added worker node.")
	ScaleClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per added worker node.")

	ScaleCommand.AddCommand(ScaleClusterCommand)
}

// scaleClusterPreRunOutput calls a pre-check function. In case anything is missing,
// displays the error and exits with code 1.
func scaleClusterPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultScaleClusterArguments()
	err := verifyScaleClusterPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	// print non-common error
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

// verifyScaleClusterPreconditions does a few general checks and returns an error in case something is missing.
func verifyScaleClusterPreconditions(args scaleClusterArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(clusterIDMissingError)
	}
	return nil
}

// scaleClusterRunOutput invokes the actual cluster scaling and prints the result and/or errors.
func scaleClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultScaleClusterArguments()
	args.clusterID = cmdLineArgs[0]

	result, err := scaleCluster(args)
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

	// Print success output
	workerWordTotal := "workers"
	if result.numWorkersAfter == 1 {
		workerWordTotal = "worker"
	}
	workerWordDiff := "workers"
	if math.Abs(float64(result.numWorkersToAdd)) == 1 {
		workerWordDiff = "worker"
	}

	if result.numWorkersToAdd > 0 {
		// scaling up
		fmt.Println(color.GreenString("The cluster is being scaled up"))
		fmt.Printf("Adding %d %s to the cluster for a total of %d %s.\n",
			result.numWorkersToAdd, workerWordDiff,
			result.numWorkersAfter, workerWordTotal)
	} else {
		// scaling down
		fmt.Println(color.GreenString("The cluster is being scaled down"))
		fmt.Printf("Removing %d %s from the cluster for a total of %d %s.\n",
			int(math.Abs(float64(result.numWorkersToAdd))), workerWordDiff,
			result.numWorkersAfter, workerWordTotal)
	}
}

// scaleCluster is the actual function submitting the API call and handling the response.
func scaleCluster(args scaleClusterArguments) (scaleClusterResults, error) {
	results := scaleClusterResults{}

	if args.numWorkersDesired == 0 {
		// here we enforce a minimum workers count of 1
		return results, microerror.Mask(cannotScaleBelowMinimumWorkersError)
	}

	clusterDetails, err := getClusterDetails(args.clusterID,
		args.scheme, args.authToken, args.apiEndpoint)
	if err != nil {
		return results, microerror.Mask(err)
	}
	results.numWorkersBefore = len(clusterDetails.Workers)
	results.numWorkersAfter = results.numWorkersBefore
	results.numWorkersToAdd = args.numWorkersDesired - results.numWorkersBefore

	if results.numWorkersToAdd == 0 {
		return results, microerror.Mask(desiredEqualsCurrentStateError)
	}

	// confirmation in case of scaling down
	if !args.oppressConfirmation && results.numWorkersToAdd < 0 {
		confirmed := askForConfirmation(fmt.Sprintf("Do you really want to reduce the worker nodes for cluster '%s' to %d?",
			args.clusterID, args.numWorkersDesired))
		if !confirmed {
			return results, microerror.Mask(commandAbortedError)
		}
	}

	// Preparing API call.
	reqBody := &models.V4ModifyClusterRequest{
		Workers: []*models.V4ModifyClusterRequestWorkersItems{},
	}
	for i := 0; i < args.numWorkersDesired; i++ {
		worker := &models.V4ModifyClusterRequestWorkersItems{}
		// worker configuration is only needed in case of scaling up,
		// but it doesn't hurt otherwise.
		if args.workerNumCPUs > 0 {
			worker.CPU.Cores = int64(args.workerNumCPUs)
		}
		if args.workerMemorySizeGB > 0 {
			worker.Memory.SizeGb = float64(args.workerMemorySizeGB)
		}
		if args.workerStorageSizeGB > 0 {
			worker.Storage.SizeGb = float64(args.workerStorageSizeGB)
		}
		reqBody.Workers = append(reqBody.Workers, worker)
	}

	// perform API call
	if args.verbose {
		fmt.Println("Sending API request to modify cluster")
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = scaleClusterActivityName

	response, err := ClientV2.ModifyCluster(args.clusterID, reqBody, auxParams)
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

	results.numWorkersAfter = len(response.Payload.Workers)

	return results, nil
}
