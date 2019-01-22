package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/juju/errgo"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
)

// addClusterArguments contains all possible input parameter needed
// (and optionally available) for creating a cluster
type addClusterArguments struct {
	apiEndpoint             string
	availabilityZones       int
	clusterName             string
	dryRun                  bool
	inputYAMLFile           string
	numWorkers              int
	owner                   string
	releaseVersion          string
	scheme                  string
	token                   string
	wokerAwsEc2InstanceType string
	wokerAzureVMSize        string
	workerNumCPUs           int
	workerMemorySizeGB      float32
	workerStorageSizeGB     float32
	workersMax              int64
	workersMin              int64
	verbose                 bool
}

func defaultAddClusterArguments() addClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)
	return addClusterArguments{
		apiEndpoint:             endpoint,
		availabilityZones:       cmdAvailabilityZones,
		clusterName:             cmdClusterName,
		dryRun:                  cmdDryRun,
		inputYAMLFile:           cmdInputYAMLFile,
		numWorkers:              cmdNumWorkers,
		owner:                   cmdOwner,
		releaseVersion:          cmdRelease,
		scheme:                  scheme,
		token:                   token,
		wokerAwsEc2InstanceType: cmdWorkerAwsEc2InstanceType,
		wokerAzureVMSize:        cmdWorkerAzureVMSize,
		workerNumCPUs:           cmdWorkerNumCPUs,
		workerMemorySizeGB:      cmdWorkerMemorySizeGB,
		workerStorageSizeGB:     cmdWorkerStorageSizeGB,
		workersMax:              cmdWorkersMax,
		workersMin:              cmdWorkersMin,
		verbose:                 cmdVerbose,
	}
}

type addClusterResult struct {
	// cluster ID
	id string
	// location to fetch details on new cluster from
	location string
	// cluster definition assembled
	definition clusterDefinition
}

const (
	// TODO: These settings should come from the API.
	// See https://github.com/giantswarm/gsctl/issues/155
	minimumNumWorkers          int     = 1
	minimumWorkerNumCPUs       int     = 1
	minimumWorkerMemorySizeGB  float32 = 1
	minimumWorkerStorageSizeGB float32 = 1

	createClusterActivityName = "create-cluster"
)

var (

	// CreateClusterCommand performs the "create cluster" function
	CreateClusterCommand = &cobra.Command{
		Use:   "cluster",
		Short: "Create cluster",
		Long: `Creates a new Kubernetes cluster.

For simple specification of a set of equal worker nodes, command line flags can
be used.

Alternatively, the --file|-f flag allows to pass a detailed definition YAML file
that can contain specs for each individual worker node, like number of CPUs,
memory size, local storage size, and node labels.

When using a definition file, some command line flags like --name|-n and
--owner|-o can be used to extend the definition given as a file. Command line
flags take precedence.

Examples:

	gsctl create cluster --file my-cluster.yaml

	gsctl create cluster -o myorg -n "My Cluster" --num-workers 5 --num-cpus 2

	gsctl create cluster -o myorg -n "My AWS Cluster" --workers-min 2 --aws-instance-type m3.medium

	gsctl create cluster -o myorg -n "My Azure Cluster" --workers-max 2 --azure-vm-size Standard_D2s_v3

	gsctl create cluster -o myorg -n "Cluster using specifc version" -r 1.2.3

	gsctl create cluster -o myorg --workers-min 3 --dry-run --verbose

	`,
		PreRun: createClusterValidationOutput,
		Run:    createClusterExecutionOutput,
	}
	cmdAvailabilityZones int
	// path to the input file used optionally as cluster definition
	cmdInputYAMLFile string
	// cluster name set via flag on execution
	cmdClusterName string
	// owner organization of the cluster as set via flag on execution
	cmdOwner string
	// number of workers required via flag on execution
	cmdNumWorkers int
	// AWS EC2 instance type to use, provided as a command line flag
	cmdWorkerAwsEc2InstanceType string
	// Azure VmSize to use, provided as a command line flag
	cmdWorkerAzureVMSize string
	// cmdWorkersMin is the minimum number of workers created for the cluster.
	cmdWorkersMin int64
	// cmdWorkersMax is the minimum number of workers created for the cluster.
	cmdWorkersMax int64
	// dry run command line flag
	cmdDryRun bool
)

func init() {
	CreateClusterCommand.Flags().IntVarP(&cmdAvailabilityZones, "availability-zones", "", 0, "Number of availability zones to use on AWS. Default is 1.")
	CreateClusterCommand.Flags().StringVarP(&cmdInputYAMLFile, "file", "f", "", "Path to a cluster definition YAML file")
	CreateClusterCommand.Flags().StringVarP(&cmdClusterName, "name", "n", "", "Cluster name")
	CreateClusterCommand.Flags().StringVarP(&cmdOwner, "owner", "o", "", "Organization to own the cluster")
	CreateClusterCommand.Flags().StringVarP(&cmdRelease, "release", "r", "", "Release version to use, e. g. '1.2.3'. Defaults to the latest. See 'gsctl list releases --help' for details.")
	CreateClusterCommand.Flags().IntVarP(&cmdNumWorkers, "num-workers", "", 0, "Number of worker nodes. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Int64VarP(&cmdWorkersMin, "workers-min", "", 0, "Minimum number of worker nodes. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Int64VarP(&cmdWorkersMax, "workers-max", "", 0, "Maximum number of worker nodes. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().StringVarP(&cmdWorkerAwsEc2InstanceType, "aws-instance-type", "", "", "EC2 instance type to use for workers (AWS only), e. g. 'm3.large'")
	CreateClusterCommand.Flags().StringVarP(&cmdWorkerAzureVMSize, "azure-vm-size", "", "", "VmSize to use for workers (Azure only), e. g. 'Standard_D2s_v3'")
	CreateClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, "num-cpus", "", 0, "Number of CPU cores per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, "memory-gb", "", 0, "RAM per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().BoolVarP(&cmdDryRun, "dry-run", "", false, "If set, the cluster won't be created. Useful with -v|--verbose.")

	// kubernetes-version never had any effect, and is deprecated now on the API side, too
	CreateClusterCommand.Flags().MarkDeprecated("kubernetes-version", "please use --release to specify a release to use")
	CreateClusterCommand.Flags().MarkDeprecated("num-workers", "please use --workers-min and --workers-max to specify the node count to use")

	CreateCommand.AddCommand(CreateClusterCommand)
}

// createClusterValidationOutput runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func createClusterValidationOutput(cmd *cobra.Command, args []string) {
	aca := defaultAddClusterArguments()

	headline := ""
	subtext := ""

	err := validateCreateClusterPreConditions(aca)
	if err != nil {
		handleCommonErrors(err)

		switch {
		case IsConflictingFlagsError(err):
			headline = "Conflicting flags used"
			subtext = "When specifying a definition via a YAML file, certain flags must not be used."
		case IsConflictingWorkerFlagsUsed(err):
			headline = "Conflicting flags used"
			subtext = "When specifying --num-workers, neither --workers-max nor --workers-min must be used."
		case IsWorkersMinMaxInvalid(err):
			headline = "Number of worker nodes invalid"
			subtext = "Node count flag --workers-min must not be higher than --workers-max."
		case IsNumWorkerNodesMissingError(err):
			headline = "Number of worker nodes required"
			subtext = "When specifying worker node details, you must also specify the number of worker nodes."
		case IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("You'll need at least %v worker nodes for a useful cluster.", minimumNumWorkers)
		case IsNotEnoughCPUCoresPerWorkerError(err):
			headline = "Not enough CPUs per worker specified"
			subtext = fmt.Sprintf("You'll need at least %v CPU cores per worker node.", minimumWorkerNumCPUs)
		case IsNotEnoughMemoryPerWorkerError(err):
			headline = "Not enough Memory per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", minimumWorkerMemorySizeGB)
		case IsNotEnoughStoragePerWorkerError(err):
			headline = "Not enough Storage per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", minimumWorkerStorageSizeGB)
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

// createClusterExecutionOutput calls addCluster() and creates user-friendly output of the result
func createClusterExecutionOutput(cmd *cobra.Command, args []string) {
	// use arguments as passed from command line via cobra
	aca := defaultAddClusterArguments()

	result, err := addCluster(aca)
	if err != nil {
		handleCommonErrors(err)

		var headline string
		var subtext string
		richError, richErrorOK := err.(*errgo.Err)

		switch {
		case IsClusterOwnerMissingError(err):
			headline = "No owner organization set"
			subtext = "Please specify an owner organization for the cluster via the --owner flag."
			if aca.inputYAMLFile != "" {
				subtext = "Please specify an owner organization for the cluster in your definition file or set one via the --owner flag."
			}
		case IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("If you specify workers in your definition file, you'll have to specify at least %d worker nodes for a useful cluster.", minimumNumWorkers)
		case IsYAMLFileNotReadableError(err):
			headline = "Could not read YAML file"
			subtext = fmt.Sprintf("The file '%s' could not read. Please make sure that it is valid YAML.", aca.inputYAMLFile)
		case IsCouldNotCreateJSONRequestBodyError(err):
			headline = "Could not create the JSON body for cluster creation API request"
			subtext = "There seems to be a problem in parsing the cluster definition. Please contact Giant Swarm via Slack or via support@giantswarm.io with details on how you executes this command."
		case IsNotAuthorizedError(err):
			headline = "Not authorized"
			subtext = "No cluster has been created, as you are are not authenticated or not authorized to perform this action."
			subtext += " Please check your credentials or, to make sure, use 'gsctl login' to log in again."
		case IsOrganizationNotFoundError(err):
			headline = "Organization not found"
			subtext = "The organization set to own the cluster does not exist."
		case IsCouldNotCreateClusterError(err):
			headline = "The cluster could not be created."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"

			// more details for backend side / connection errors
			subtext += "\n\nDetails:\n"
			if richErrorOK {
				subtext += richError.Message()
			} else {
				subtext += err.Error()
			}

		default:
			headline = err.Error()
		}

		// output error information
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	// success output
	if !aca.dryRun {
		if result.definition.Name != "" {
			fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' is launching.", result.definition.Name, result.id, result.definition.Owner))
		} else {
			fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' is launching.", result.id, result.definition.Owner))
		}
		fmt.Println("Add key pair and settings to kubectl using")
		fmt.Println("")
		fmt.Printf("    %s", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s \n", result.id)))
		fmt.Println("")
		fmt.Println("Take into consideration all clusters have enabled RBAC and may you want to provide a correct organization for the certificates (like operators, testers, developer, ...)")
		fmt.Println("")
		fmt.Printf("    %s \n", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s --certificate-organizations system:masters", result.id)))
		fmt.Println("")
		fmt.Println("To know more about how to create the kubeconfig run")
		fmt.Println("")
		fmt.Printf("    %s \n\n", color.YellowString("gsctl create kubeconfig --help"))
	}
}

// validateCreateClusterPreConditions checks preconditions and returns an error in case
func validateCreateClusterPreConditions(args addClusterArguments) error {
	// logged in?
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}

	// false flag combination?
	if args.inputYAMLFile != "" {
		if args.numWorkers != 0 || args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0 || args.wokerAwsEc2InstanceType != "" || args.wokerAzureVMSize != "" {
			return microerror.Mask(conflictingFlagsError)
		}
	}

	// validate number of workers specified by flag
	if args.numWorkers > 0 && (args.workersMax > 0 || args.workersMin > 0) {
		return microerror.Mask(conflictingWorkerFlagsUsedError)
	}
	if args.numWorkers > 0 && args.numWorkers < minimumNumWorkers {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMax > 0 && args.workersMax < int64(minimumNumWorkers) {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMin > 0 && args.workersMin < int64(minimumNumWorkers) {
		return microerror.Mask(notEnoughWorkerNodesError)
	}
	if args.workersMin > 0 && args.workersMax > 0 && args.workersMin > args.workersMax {
		return microerror.Mask(workersMinMaxInvalidError)
	}

	// validate number of CPUs specified by flag
	if args.workerNumCPUs > 0 && args.workerNumCPUs < minimumWorkerNumCPUs {
		return microerror.Mask(notEnoughCPUCoresPerWorkerError)
	}

	// validate memory size specified by flag
	if args.workerMemorySizeGB > 0 && args.workerMemorySizeGB < minimumWorkerMemorySizeGB {
		return microerror.Mask(notEnoughMemoryPerWorkerError)
	}

	// validate storage size specified by flag
	if args.workerStorageSizeGB > 0 && args.workerStorageSizeGB < minimumWorkerStorageSizeGB {
		return microerror.Mask(notEnoughStoragePerWorkerError)
	}

	if args.wokerAwsEc2InstanceType != "" || args.wokerAzureVMSize != "" {
		// check for incompatibilities
		if args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0 {
			return microerror.Mask(incompatibleSettingsError)
		}
	}

	return nil
}

// readDefinitionFromFile reads a cluster definition from a YAML config file
func readDefinitionFromFile(path string) (clusterDefinition, error) {
	def := clusterDefinition{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return clusterDefinition{}, microerror.Mask(err)
	}

	err = yaml.Unmarshal(data, &def)
	if err != nil {
		return clusterDefinition{}, microerror.Mask(err)
	}

	return def, nil
}

// createDefinitionFromFlags creates a clusterDefinition based on the
// flags/arguments the user has given
func definitionFromFlags(def clusterDefinition, args addClusterArguments) clusterDefinition {
	if args.availabilityZones != 0 {
		def.AvailabilityZones = args.availabilityZones
	}

	if args.clusterName != "" {
		def.Name = args.clusterName
	}

	if args.releaseVersion != "" {
		def.ReleaseVersion = args.releaseVersion
	}

	if def.Scaling.Min > 0 && args.workersMin == 0 {
		args.workersMin = def.Scaling.Min
	}
	if def.Scaling.Max > 0 && args.workersMax == 0 {
		args.workersMax = def.Scaling.Max
	}

	if args.workersMax > 0 {
		def.Scaling.Max = args.workersMax
		args.numWorkers = 1
		if args.workersMin == 0 {
			def.Scaling.Min = def.Scaling.Max
		}
	}
	if args.workersMin > 0 {
		def.Scaling.Min = args.workersMin
		args.numWorkers = 1
		if args.workersMax == 0 {
			def.Scaling.Max = def.Scaling.Min
		}
	}

	if args.owner != "" {
		def.Owner = args.owner
	}

	if args.numWorkers != 0 {
		workers := []nodeDefinition{}

		for i := 0; i < args.numWorkers; i++ {
			worker := nodeDefinition{}

			if args.workerNumCPUs != 0 {
				worker.CPU = cpuDefinition{Cores: args.workerNumCPUs}
			}

			if args.workerStorageSizeGB != 0 {
				worker.Storage = storageDefinition{SizeGB: args.workerStorageSizeGB}
			}

			if args.workerMemorySizeGB != 0 {
				worker.Memory = memoryDefinition{SizeGB: args.workerMemorySizeGB}
			}

			// AWS-specific
			if args.wokerAwsEc2InstanceType != "" {
				worker.AWS.InstanceType = args.wokerAwsEc2InstanceType
			}

			// Azure
			if args.wokerAzureVMSize != "" {
				worker.Azure.VMSize = args.wokerAzureVMSize
			}

			workers = append(workers, worker)
		}

		def.Workers = workers
		if def.Scaling.Min == 0 && def.Scaling.Max == 0 {
			def.Scaling.Min = int64(len(def.Workers))
			def.Scaling.Max = int64(len(def.Workers))
		}
	}

	return def
}

// creates a models.V4AddClusterRequest from clusterDefinition
func createAddClusterBody(d clusterDefinition) *models.V4AddClusterRequest {
	a := &models.V4AddClusterRequest{}
	a.AvailabilityZones = int64(d.AvailabilityZones)
	a.Name = d.Name
	a.Owner = &d.Owner
	a.ReleaseVersion = d.ReleaseVersion
	a.Scaling = &models.V4AddClusterRequestScaling{
		Min: d.Scaling.Min,
		Max: d.Scaling.Max,
	}

	for _, dWorker := range d.Workers {
		ndmWorker := &models.V4AddClusterRequestWorkersItems{}
		ndmWorker.Memory = &models.V4AddClusterRequestWorkersItemsMemory{SizeGb: float64(dWorker.Memory.SizeGB)}
		ndmWorker.CPU = &models.V4AddClusterRequestWorkersItemsCPU{Cores: int64(dWorker.CPU.Cores)}
		ndmWorker.Storage = &models.V4AddClusterRequestWorkersItemsStorage{SizeGb: float64(dWorker.Storage.SizeGB)}
		ndmWorker.Labels = dWorker.Labels
		ndmWorker.Aws = &models.V4AddClusterRequestWorkersItemsAws{InstanceType: dWorker.AWS.InstanceType}
		ndmWorker.Azure = &models.V4AddClusterRequestWorkersItemsAzure{VMSize: dWorker.Azure.VMSize}
		a.Workers = append(a.Workers, ndmWorker)
	}

	return a
}

// addCluster actually adds a cluster, interpreting all the input Configuration
// and returning a structured result
func addCluster(args addClusterArguments) (addClusterResult, error) {
	var result addClusterResult
	var err error

	if args.inputYAMLFile != "" {
		// definition from file (and optionally flags)
		result.definition, err = readDefinitionFromFile(args.inputYAMLFile)
		if err != nil {
			return addClusterResult{}, microerror.Maskf(yamlFileNotReadableError, err.Error())
		}
		result.definition = definitionFromFlags(result.definition, args)
	} else {
		// definition from flags only
		result.definition = definitionFromFlags(clusterDefinition{}, args)
	}

	// Validate definition
	if result.definition.Owner == "" {
		return addClusterResult{}, microerror.Mask(clusterOwnerMissingError)
	}

	// Validations based on definition file.
	// For validations based on command line flags, see validateCreateClusterPreConditions()
	if args.inputYAMLFile != "" {
		// number of workers
		if len(result.definition.Workers) > 0 && len(result.definition.Workers) < minimumNumWorkers {
			return addClusterResult{}, microerror.Mask(notEnoughWorkerNodesError)
		}
	}

	// Preview in YAML format
	if args.verbose {
		fmt.Println("\nDefinition for the requested cluster:")
		d, marshalErr := yaml.Marshal(result.definition)
		if marshalErr != nil {
			log.Fatalf("error: %v", marshalErr)
		}
		fmt.Printf(color.CyanString(string(d)))
		fmt.Println()
	}

	// create JSON API call payload to catch and handle errors early
	addClusterBody := createAddClusterBody(result.definition)
	_, marshalErr := json.Marshal(addClusterBody)
	if marshalErr != nil {
		return result, microerror.Maskf(couldNotCreateJSONRequestBodyError, marshalErr.Error())
	}

	if !args.dryRun {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(result.definition.Owner))

		auxParams := ClientV2.DefaultAuxiliaryParams()
		auxParams.ActivityName = createClusterActivityName

		// perform API call
		response, err := ClientV2.CreateCluster(addClusterBody, auxParams)
		if err != nil {
			// create specific error types for cases we care about
			if clientErr, ok := err.(*clienterror.APIError); ok {
				if clientErr.HTTPStatusCode == http.StatusNotFound {
					// owner org not existing
					return result, microerror.Mask(organizationNotFoundError)
				} else if clientErr.HTTPStatusCode == http.StatusUnauthorized {
					// not authorized
					return result, microerror.Mask(notAuthorizedError)
				}
			}

			return result, microerror.Mask(err)
		}

		// success

		result.location = response.Location
		result.id = strings.Split(result.location, "/")[3]
	}

	return result, nil

}
