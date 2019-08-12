// Package cluster defines the 'create cluster' command.
package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/juju/errgo"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/limits"
)

// Arguments contains all possible input parameter needed
// (and optionally available) for creating a cluster
type Arguments struct {
	apiEndpoint             string
	availabilityZones       int
	clusterName             string
	dryRun                  bool
	inputYAMLFile           string
	fileSystem              afero.Fs
	numWorkers              int
	owner                   string
	releaseVersion          string
	scheme                  string
	token                   string
	userProvidedToken       string
	verbose                 bool
	wokerAwsEc2InstanceType string
	wokerAzureVMSize        string
	workerMemorySizeGB      float32
	workerNumCPUs           int
	workersMax              int64
	workersMin              int64
	workerStorageSizeGB     float32
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:             endpoint,
		availabilityZones:       cmdAvailabilityZones,
		clusterName:             cmdClusterName,
		dryRun:                  cmdDryRun,
		inputYAMLFile:           cmdInputYAMLFile,
		fileSystem:              config.FileSystem,
		numWorkers:              flags.NumWorkers,
		owner:                   cmdOwner,
		releaseVersion:          flags.Release,
		scheme:                  scheme,
		token:                   token,
		userProvidedToken:       flags.Token,
		verbose:                 flags.Verbose,
		wokerAwsEc2InstanceType: flags.WorkerAwsEc2InstanceType,
		wokerAzureVMSize:        cmdWorkerAzureVMSize,
		workerMemorySizeGB:      flags.WorkerMemorySizeGB,
		workerNumCPUs:           flags.WorkerNumCPUs,
		workersMax:              flags.WorkersMax,
		workersMin:              flags.WorkersMin,
		workerStorageSizeGB:     flags.WorkerStorageSizeGB,
	}
}

type creationResult struct {
	// cluster ID
	id string
	// location to fetch details on new cluster from
	location string
	// cluster definition assembled
	definition types.ClusterDefinition
}

const (
	createClusterActivityName = "create-cluster"
)

var (

	// Command performs the "create cluster" function
	Command = &cobra.Command{
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

  gsctl create cluster \
    -o myorg -n "My KVM Cluster" \
    --num-workers 5 --num-cpus 2

  gsctl create cluster \
    -o myorg -n "My AWS Autoscaling Cluster" \
    --workers-min 3 --workers-max 6 \
    --aws-instance-type m3.xlarge

  gsctl create cluster \
    -o myorg -n "My Azure Cluster" \
    --num-workers 5 \
    --azure-vm-size Standard_D2s_v3

  gsctl create cluster \
    -o myorg -n "Cluster using specific version" \
    --release 1.2.3

  gsctl create cluster \
    -o myorg --num-workers 3 \
    --dry-run --verbose

`,
		PreRun: printValidation,
		Run:    printResult,
	}
	cmdAvailabilityZones int
	// path to the input file used optionally as cluster definition
	cmdInputYAMLFile string
	// cluster name set via flag on execution
	cmdClusterName string
	// owner organization of the cluster as set via flag on execution
	cmdOwner string
	// Azure VmSize to use, provided as a command line flag
	cmdWorkerAzureVMSize string
	// dry run command line flag
	cmdDryRun bool
)

func init() {
	Command.Flags().IntVarP(&cmdAvailabilityZones, "availability-zones", "", 0, "Number of availability zones to use on AWS. Default is 1.")
	Command.Flags().StringVarP(&cmdInputYAMLFile, "file", "f", "", "Path to a cluster definition YAML file")
	Command.Flags().StringVarP(&cmdClusterName, "name", "n", "", "Cluster name")
	Command.Flags().StringVarP(&cmdOwner, "owner", "o", "", "Organization to own the cluster")
	Command.Flags().StringVarP(&flags.Release, "release", "r", "", "Release version to use, e. g. '1.2.3'. Defaults to the latest. See 'gsctl list releases --help' for details.")
	Command.Flags().IntVarP(&flags.NumWorkers, "num-workers", "", 0, "Shorthand to set --workers-min and --workers-max to the same value. Can't be used with -f|--file.")
	Command.Flags().Int64VarP(&flags.WorkersMin, "workers-min", "", 0, "Minimum number of worker nodes. Can't be used with -f|--file.")
	Command.Flags().Int64VarP(&flags.WorkersMax, "workers-max", "", 0, "Maximum number of worker nodes. Can't be used with -f|--file.")
	Command.Flags().StringVarP(&flags.WorkerAwsEc2InstanceType, "aws-instance-type", "", "", "EC2 instance type to use for workers (AWS only), e. g. 'm3.large'")
	Command.Flags().StringVarP(&cmdWorkerAzureVMSize, "azure-vm-size", "", "", "VmSize to use for workers (Azure only), e. g. 'Standard_D2s_v3'")
	Command.Flags().IntVarP(&flags.WorkerNumCPUs, "num-cpus", "", 0, "Number of CPU cores per worker node. Can't be used with -f|--file.")
	Command.Flags().Float32VarP(&flags.WorkerMemorySizeGB, "memory-gb", "", 0, "RAM per worker node. Can't be used with -f|--file.")
	Command.Flags().Float32VarP(&flags.WorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per worker node. Can't be used with -f|--file.")
	Command.Flags().BoolVarP(&cmdDryRun, "dry-run", "", false, "If set, the cluster won't be created. Useful with -v|--verbose.")
}

// printValidation runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func printValidation(cmd *cobra.Command, args []string) {
	aca := collectArguments()

	headline := ""
	subtext := ""

	err := validatePreConditions(aca)
	if err != nil {
		errors.HandleCommonErrors(err)

		switch {
		case errors.IsConflictingFlagsError(err):
			headline = "Conflicting flags used"
			subtext = "When specifying a definition via a YAML file, certain flags must not be used."
		case errors.IsConflictingWorkerFlagsUsed(err):
			headline = "Conflicting flags used"
			subtext = "When specifying --num-workers, neither --workers-max nor --workers-min must be used."
		case errors.IsWorkersMinMaxInvalid(err):
			headline = "Number of worker nodes invalid"
			subtext = "Node count flag --workers-min must not be higher than --workers-max."
		case errors.IsNumWorkerNodesMissingError(err):
			headline = "Number of worker nodes required"
			subtext = "When specifying worker node details, you must also specify the number of worker nodes."
		case errors.IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("You'll need at least %v worker nodes for a useful cluster.", limits.MinimumNumWorkers)
		case errors.IsNotEnoughCPUCoresPerWorkerError(err):
			headline = "Not enough CPUs per worker specified"
			subtext = fmt.Sprintf("You'll need at least %v CPU cores per worker node.", limits.MinimumWorkerNumCPUs)
		case errors.IsNotEnoughMemoryPerWorkerError(err):
			headline = "Not enough Memory per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", limits.MinimumWorkerMemorySizeGB)
		case errors.IsNotEnoughStoragePerWorkerError(err):
			headline = "Not enough Storage per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", limits.MinimumWorkerStorageSizeGB)
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

// printResult calls addCluster() and creates user-friendly output of the result
func printResult(cmd *cobra.Command, args []string) {
	// use arguments as passed from command line via cobra
	aca := collectArguments()

	result, err := addCluster(aca)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		var headline string
		var subtext string
		richError, richErrorOK := err.(*errgo.Err)

		switch {
		case errors.IsClusterOwnerMissingError(err):
			headline = "No owner organization set"
			subtext = "Please specify an owner organization for the cluster via the --owner flag."
			if aca.inputYAMLFile != "" {
				subtext = "Please specify an owner organization for the cluster in your definition file or set one via the --owner flag."
			}
		case errors.IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("If you specify workers in your definition file, you'll have to specify at least %d worker nodes for a useful cluster.", limits.MinimumNumWorkers)
		case errors.IsYAMLFileNotReadableError(err):
			headline = "Could not read YAML file"
			subtext = fmt.Sprintf("The file '%s' could not read. Please make sure that it is valid YAML.", aca.inputYAMLFile)
		case errors.IsCouldNotCreateJSONRequestBodyError(err):
			headline = "Could not create the JSON body for cluster creation API request"
			subtext = "There seems to be a problem in parsing the cluster definition. Please contact Giant Swarm via Slack or via support@giantswarm.io with details on how you executes this command."
		case errors.IsNotAuthorizedError(err):
			headline = "Not authorized"
			subtext = "No cluster has been created, as you are are not authenticated or not authorized to perform this action."
			subtext += " Please check your credentials or, to make sure, use 'gsctl login' to log in again."
		case errors.IsOrganizationNotFoundError(err):
			headline = "Organization not found"
			subtext = "The organization set to own the cluster does not exist."
		case errors.IsCouldNotCreateClusterError(err):
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

// validatePreConditions checks preconditions and returns an error in case
func validatePreConditions(args Arguments) error {
	// logged in?
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	// false flag combination?
	if args.inputYAMLFile != "" {
		if args.numWorkers != 0 || args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0 || args.wokerAwsEc2InstanceType != "" || args.wokerAzureVMSize != "" {
			return microerror.Mask(errors.ConflictingFlagsError)
		}
	}

	// validate number of workers specified by flag
	if args.numWorkers > 0 && (args.workersMax > 0 || args.workersMin > 0) {
		return microerror.Mask(errors.ConflictingWorkerFlagsUsedError)
	}
	if args.numWorkers > 0 && args.numWorkers < limits.MinimumNumWorkers {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.workersMax > 0 && args.workersMax < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.workersMin > 0 && args.workersMin < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.workersMin > 0 && args.workersMax > 0 && args.workersMin > args.workersMax {
		return microerror.Mask(errors.WorkersMinMaxInvalidError)
	}

	// validate number of CPUs specified by flag
	if args.workerNumCPUs > 0 && args.workerNumCPUs < limits.MinimumWorkerNumCPUs {
		return microerror.Mask(errors.NotEnoughCPUCoresPerWorkerError)
	}

	// validate memory size specified by flag
	if args.workerMemorySizeGB > 0 && args.workerMemorySizeGB < limits.MinimumWorkerMemorySizeGB {
		return microerror.Mask(errors.NotEnoughMemoryPerWorkerError)
	}

	// validate storage size specified by flag
	if args.workerStorageSizeGB > 0 && args.workerStorageSizeGB < limits.MinimumWorkerStorageSizeGB {
		return microerror.Mask(errors.NotEnoughStoragePerWorkerError)
	}

	if args.wokerAwsEc2InstanceType != "" || args.wokerAzureVMSize != "" {
		// check for incompatibilities
		if args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0 {
			return microerror.Mask(errors.IncompatibleSettingsError)
		}
	}

	return nil
}

// readDefinitionFromFile reads a cluster definition from a YAML config file
func readDefinitionFromFile(fs afero.Fs, path string) (types.ClusterDefinition, error) {
	def := types.ClusterDefinition{}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return types.ClusterDefinition{}, microerror.Mask(err)
	}

	err = yaml.Unmarshal(data, &def)
	if err != nil {
		return types.ClusterDefinition{}, microerror.Mask(err)
	}

	return def, nil
}

// createDefinitionFromFlags creates a clusterDefinition based on the
// flags/arguments the user has given
func definitionFromFlags(def types.ClusterDefinition, args Arguments) types.ClusterDefinition {
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

	if def.Scaling.Min == 0 && def.Scaling.Max == 0 {
		def.Scaling.Min = int64(args.numWorkers)
		def.Scaling.Max = int64(args.numWorkers)
	}

	if def.Scaling.Min == 0 && def.Scaling.Max == 0 && args.numWorkers == 0 {
		def.Scaling.Min = 3
		def.Scaling.Max = 3
	}

	workers := []types.NodeDefinition{}

	worker := types.NodeDefinition{}
	if args.workerNumCPUs != 0 {
		worker.CPU = types.CPUDefinition{Cores: args.workerNumCPUs}
	}
	if args.workerStorageSizeGB != 0 {
		worker.Storage = types.StorageDefinition{SizeGB: args.workerStorageSizeGB}
	}
	if args.workerMemorySizeGB != 0 {
		worker.Memory = types.MemoryDefinition{SizeGB: args.workerMemorySizeGB}
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

	def.Workers = workers

	return def
}

// creates a models.V4AddClusterRequest from clusterDefinition
func createAddClusterBody(d types.ClusterDefinition) *models.V4AddClusterRequest {
	a := &models.V4AddClusterRequest{}
	a.AvailabilityZones = int64(d.AvailabilityZones)
	a.Name = d.Name
	a.Owner = &d.Owner
	a.ReleaseVersion = d.ReleaseVersion
	a.Scaling = &models.V4AddClusterRequestScaling{
		Min: d.Scaling.Min,
		Max: d.Scaling.Max,
	}

	if len(d.Workers) == 1 {
		ndmWorker := &models.V4AddClusterRequestWorkersItems{}
		ndmWorker.Memory = &models.V4AddClusterRequestWorkersItemsMemory{SizeGb: float64(d.Workers[0].Memory.SizeGB)}
		ndmWorker.CPU = &models.V4AddClusterRequestWorkersItemsCPU{Cores: int64(d.Workers[0].CPU.Cores)}
		ndmWorker.Storage = &models.V4AddClusterRequestWorkersItemsStorage{SizeGb: float64(d.Workers[0].Storage.SizeGB)}
		ndmWorker.Labels = d.Workers[0].Labels
		ndmWorker.Aws = &models.V4AddClusterRequestWorkersItemsAws{InstanceType: d.Workers[0].AWS.InstanceType}
		ndmWorker.Azure = &models.V4AddClusterRequestWorkersItemsAzure{VMSize: d.Workers[0].Azure.VMSize}
		a.Workers = append(a.Workers, ndmWorker)
	}

	return a
}

// addCluster actually adds a cluster, interpreting all the input Configuration
// and returning a structured result
func addCluster(args Arguments) (creationResult, error) {
	var result creationResult
	var err error

	if args.inputYAMLFile != "" {
		// definition from file (and optionally flags)
		result.definition, err = readDefinitionFromFile(args.fileSystem, args.inputYAMLFile)
		if err != nil {
			return creationResult{}, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
		}
		result.definition = definitionFromFlags(result.definition, args)
	} else {
		// definition from flags only
		result.definition = definitionFromFlags(types.ClusterDefinition{}, args)
	}

	// Validate definition
	if result.definition.Owner == "" {
		return creationResult{}, microerror.Mask(errors.ClusterOwnerMissingError)
	}

	// Validations based on definition file.
	// For validations based on command line flags, see validatePreConditions()
	if args.inputYAMLFile != "" {
		// number of workers
		if len(result.definition.Workers) > 0 && len(result.definition.Workers) < limits.MinimumNumWorkers {
			return creationResult{}, microerror.Mask(errors.NotEnoughWorkerNodesError)
		}
	}

	// create JSON API call payload to catch and handle errors early
	addClusterBody := createAddClusterBody(result.definition)
	_, marshalErr := json.Marshal(addClusterBody)
	if marshalErr != nil {
		return result, microerror.Maskf(errors.CouldNotCreateJSONRequestBodyError, marshalErr.Error())
	}

	// Preview in YAML format
	if args.verbose {
		fmt.Println("\nDefinition for the requested cluster:")
		d, marshalErr := yaml.Marshal(addClusterBody)
		if marshalErr != nil {
			log.Fatalf("error: %v", marshalErr)
		}
		fmt.Printf(color.CyanString(string(d)))
		fmt.Println()
	}

	if !args.dryRun {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(result.definition.Owner))

		clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
		if err != nil {
			return result, microerror.Mask(err)
		}

		auxParams := clientWrapper.DefaultAuxiliaryParams()
		auxParams.ActivityName = createClusterActivityName
		// perform API call
		response, err := clientWrapper.CreateCluster(addClusterBody, auxParams)
		if err != nil {
			return result, microerror.Mask(err)
		}

		// success
		result.location = response.Location
		result.id = strings.Split(result.location, "/")[3]
	}

	return result, nil

}
