package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	microerror "github.com/giantswarm/microkit/error"
	yaml "gopkg.in/yaml.v2"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

// addClusterArguments contains all possible input parameter needed
// (and optionally available) for creating a cluster
type addClusterArguments struct {
	apiEndpoint         string
	clusterName         string
	dryRun              bool
	inputYAMLFile       string
	kubernetesVersion   string
	numWorkers          int
	owner               string
	token               string
	workerNumCPUs       int
	workerMemorySizeGB  float32
	workerStorageSizeGB float32
	verbose             bool
}

func defaultAddClusterArguments() addClusterArguments {
	return addClusterArguments{
		apiEndpoint:         cmdAPIEndpoint,
		clusterName:         cmdClusterName,
		dryRun:              cmdDryRun,
		inputYAMLFile:       cmdInputYAMLFile,
		kubernetesVersion:   cmdKubernetesVersion,
		numWorkers:          cmdNumWorkers,
		owner:               cmdOwner,
		token:               cmdToken,
		workerNumCPUs:       cmdWorkerNumCPUs,
		workerMemorySizeGB:  cmdWorkerMemorySizeGB,
		workerStorageSizeGB: cmdWorkerStorageSizeGB,
		verbose:             cmdVerbose,
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
	// TODO: These settings should come from the API
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

When using a definition file, some command line flags like --name and --owner
can be used to extend the definition given as a file. Command line flags take
precedence.

Examples:

	gsctl create cluster --file my-cluster.yaml

	gsctl create cluster --owner=myorg --name="My Cluster" --num-workers=5 --num-cpus=2

  gsctl create cluster --owner=myorg --num-workers=3 --dry-run --verbose`,
		PreRun: validationOutput,
		Run:    executionOutput,
	}

	// path to the input file used optionally as cluster definition
	cmdInputYAMLFile string
	// cluster name set via flag on execution
	cmdClusterName string
	// Kubernetes version number required via flag on execution
	cmdKubernetesVersion string
	// owner organization of the cluster as set via flag on execution
	cmdOwner string
	// number of workers required via flag on execution
	cmdNumWorkers int
	// dry run command line flag
	cmdDryRun bool
)

func init() {
	CreateClusterCommand.Flags().StringVarP(&cmdInputYAMLFile, "file", "f", "", "Path to a cluster definition YAML file")
	CreateClusterCommand.Flags().StringVarP(&cmdClusterName, "name", "", "", "Cluster name")
	CreateClusterCommand.Flags().StringVarP(&cmdKubernetesVersion, "kubernetes-version", "", "", "Kubernetes version of the cluster")
	CreateClusterCommand.Flags().StringVarP(&cmdOwner, "owner", "", "", "Organization to own the cluster")
	CreateClusterCommand.Flags().IntVarP(&cmdNumWorkers, "num-workers", "", 0, "Number of worker nodes. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().IntVarP(&cmdWorkerNumCPUs, "num-cpus", "", 0, "Number of CPU cores per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Float32VarP(&cmdWorkerMemorySizeGB, "memory-gb", "", 0, "RAM per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().Float32VarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().BoolVarP(&cmdDryRun, "dry-run", "", false, "If set, the cluster won't be created. Useful with -v|--verbose.")

	CreateCommand.AddCommand(CreateClusterCommand)
}

// validationOutput runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func validationOutput(cmd *cobra.Command, args []string) {
	aca := defaultAddClusterArguments()

	headline := ""
	subtext := ""

	err := validatePreConditions(aca)
	if err != nil {
		switch {
		case err.Error() == "":
			return
		case IsNotLoggedInError(err):
			headline = "You are not logged in."
			subtext = fmt.Sprintf("Use '%s login' to login or '--auth-token' to pass a valid auth token.", config.ProgramName)
		case IsConflictingFlagsError(err):
			headline = "Conflicting flags used"
			subtext = "When specifying a definition via a YAML file, certain flags must not be used."
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

// executionOutput calls addCluster() and creates user-friendly output of the result
func executionOutput(cmd *cobra.Command, args []string) {
	// use arguments as passed from command line via cobra
	aca := defaultAddClusterArguments()

	result, err := addCluster(aca)
	if err != nil {
		var headline string
		var subtext string

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
		case IsCouldNotCreateClusterError(err):
			headline = "The cluster could not be created."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"
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
	if result.definition.Name != "" {
		fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' is launching.", result.definition.Name, result.id, result.definition.Owner))
	} else {
		fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' is launching.", result.id, result.definition.Owner))
	}
	fmt.Println("Add key pair and settings to kubectl using")
	fmt.Println("")
	fmt.Printf("    %s\n\n", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s", result.id)))
}

// validatePreConditions checks preconditions and returns an error in case
func validatePreConditions(args addClusterArguments) error {
	// logged in?
	if config.Config.Token == "" && args.token == "" {
		return microerror.MaskAny(notLoggedInError)
	}

	// false flag combination?
	if args.inputYAMLFile != "" {
		if args.numWorkers != 0 || args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0 {
			return microerror.MaskAny(conflictingFlagsError)
		}
	} else {
		if args.numWorkers == 0 && (args.workerNumCPUs != 0 || args.workerMemorySizeGB != 0 || args.workerStorageSizeGB != 0) {
			return microerror.MaskAny(numWorkerNodesMissingError)
		}
	}

	// validate number of workers specified by flag
	if args.numWorkers > 0 && args.numWorkers < minimumNumWorkers {
		return microerror.MaskAny(notEnoughWorkerNodesError)
	}

	// validate number of CPUs specified by flag
	if args.workerNumCPUs > 0 && args.workerNumCPUs < minimumWorkerNumCPUs {
		return microerror.MaskAny(notEnoughCPUCoresPerWorkerError)
	}

	// validate memory size specified by flag
	if args.workerMemorySizeGB > 0 && args.workerMemorySizeGB < minimumWorkerMemorySizeGB {
		return microerror.MaskAny(notEnoughMemoryPerWorkerError)
	}

	// validate storage size specified by flag
	if args.workerStorageSizeGB > 0 && args.workerStorageSizeGB < minimumWorkerStorageSizeGB {
		return microerror.MaskAny(notEnoughStoragePerWorkerError)
	}

	return nil
}

// readDefinitionFromFile reads a cluster definition from a YAML config file
func readDefinitionFromFile(filePath string) (clusterDefinition, error) {
	myDef := clusterDefinition{}
	data, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		return myDef, readErr
	}
	return unmarshalDefinition(data, myDef)
}

// unmarshalDefinition takes YAML input and returns it as a clusterDefinition
func unmarshalDefinition(data []byte, myDef clusterDefinition) (clusterDefinition, error) {
	yamlErr := yaml.Unmarshal(data, &myDef)
	if yamlErr != nil {
		return myDef, yamlErr
	}
	return myDef, nil
}

// enhanceDefinitionWithFlags takes a definition specified by file and
// overwrites some settings given via flags.
// Note that only a few attributes can be overridden by flags.
func enhanceDefinitionWithFlags(def *clusterDefinition, args addClusterArguments) {
	if args.clusterName != "" {
		def.Name = args.clusterName
	}
	if args.kubernetesVersion != "" {
		def.KubernetesVersion = args.kubernetesVersion
	}
	if args.owner != "" {
		def.Owner = args.owner
	}
}

// createDefinitionFromFlags creates a clusterDefinition based on the
// flags/arguments the user has given
func createDefinitionFromFlags(args addClusterArguments) clusterDefinition {
	def := clusterDefinition{}
	if args.clusterName != "" {
		def.Name = args.clusterName
	}
	if args.kubernetesVersion != "" {
		def.KubernetesVersion = args.kubernetesVersion
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
			workers = append(workers, worker)
		}
		def.Workers = workers
	}
	return def
}

// creates a gsclientgen.V4AddClusterRequest from clusterDefinition
func createAddClusterBody(d clusterDefinition) gsclientgen.V4AddClusterRequest {
	a := gsclientgen.V4AddClusterRequest{}
	a.Name = d.Name
	a.Owner = d.Owner
	a.KubernetesVersion = d.KubernetesVersion

	for _, dWorker := range d.Workers {
		ndmWorker := gsclientgen.V4NodeDefinition{}
		ndmWorker.Memory = gsclientgen.V4NodeDefinitionMemory{SizeGb: dWorker.Memory.SizeGB}
		ndmWorker.Cpu = gsclientgen.V4NodeDefinitionCpu{Cores: int32(dWorker.CPU.Cores)}
		ndmWorker.Storage = gsclientgen.V4NodeDefinitionStorage{SizeGb: dWorker.Storage.SizeGB}
		ndmWorker.Labels = dWorker.Labels
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
			return addClusterResult{}, microerror.MaskAnyf(yamlFileNotReadableError, err.Error())
		}
		enhanceDefinitionWithFlags(&result.definition, args)
	} else {
		// definition from flags only
		result.definition = createDefinitionFromFlags(args)
	}

	// Validate definition
	if result.definition.Owner == "" {
		return result, microerror.MaskAny(clusterOwnerMissingError)
	}

	// Validations based on definition file.
	// For validations based on command line flags, see validatePreConditions()
	if args.inputYAMLFile != "" {
		// number of workers
		if len(result.definition.Workers) > 0 && len(result.definition.Workers) < minimumNumWorkers {
			return result, microerror.MaskAny(notEnoughWorkerNodesError)
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
		return result, microerror.MaskAnyf(couldNotCreateJSONRequestBodyError, marshalErr.Error())
	}

	if !args.dryRun {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(result.definition.Owner))

		// perform API call
		authHeader := "giantswarm " + config.Config.Token
		if args.token != "" {
			// command line flag overwrites
			authHeader = "giantswarm " + args.token
		}
		clientConfig := client.Configuration{
			Endpoint:  args.apiEndpoint,
			UserAgent: config.UserAgent(),
		}
		apiClient, clientErr := client.NewClient(clientConfig)
		if clientErr != nil {
			return result, microerror.MaskAny(couldNotCreateClientError)
		}
		responseBody, apiResponse, err := apiClient.AddCluster(authHeader, addClusterBody, requestIDHeader, createClusterActivityName, cmdLine)
		if err != nil {
			// lower level connection problem
			return result, microerror.MaskAny(err)
		}

		// handle API result
		if responseBody.Code != "RESOURCE_CREATED" {
			return result, microerror.MaskAnyf(couldNotCreateClusterError,
				fmt.Sprintf("Error in API request to create cluster: %s (Code: %s)",
					responseBody.Message, responseBody.Code))
		}
		result.location = apiResponse.Header["Location"][0]
		result.id = strings.Split(result.location, "/")[3]
	}

	return result, nil

}
