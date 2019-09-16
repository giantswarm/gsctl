/*
Package cluster defines the 'create cluster' command.

The command deals with a few delicacies/spiecialties:

- Cluster spec details, e. g. which instance type to use for workers, can
  be specified by the user, but don't have to. The backend will
  fill in missing details using defaulting.

- Cluster spec details can be provided either using command line flags
  or by passing a YAML definition. When passing a YAML definition, some
  attributes from that definition can even be overridden using flags.

- The command features a dry-run mode, which is enabled using a flag. In
  this mode the resulting cluster definition is printed, but the final
  API call for cluster creation is not actually submitted.

- On AWS, starting from a certain release version, clusters will have
  node pools and will be created using the v5 API endpoint. On other providers
  as well as on AWS for older releases, the v4 API endpoint has to be used.

*/
package cluster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
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
// (and optionally available) for creating a cluster.
type Arguments struct {
	APIEndpoint              string
	AuthToken                string
	AvailabilityZones        int
	ClusterName              string
	DryRun                   bool
	FileSystem               afero.Fs
	InputYAMLFile            string
	NumWorkers               int
	Owner                    string
	ReleaseVersion           string
	Scheme                   string
	UserProvidedToken        string
	Verbose                  bool
	WorkerAwsEc2InstanceType string
	WorkerAzureVMSize        string
	WorkerMemorySizeGB       float32
	WorkerNumCPUs            int
	WorkersMax               int64
	WorkersMin               int64
	WorkerStorageSizeGB      float32
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		APIEndpoint:              endpoint,
		AuthToken:                token,
		AvailabilityZones:        cmdAvailabilityZones,
		ClusterName:              cmdClusterName,
		DryRun:                   cmdDryRun,
		FileSystem:               config.FileSystem,
		InputYAMLFile:            cmdInputYAMLFile,
		NumWorkers:               flags.NumWorkers,
		Owner:                    cmdOwner,
		ReleaseVersion:           flags.Release,
		Scheme:                   scheme,
		UserProvidedToken:        flags.Token,
		Verbose:                  flags.Verbose,
		WorkerAwsEc2InstanceType: flags.WorkerAwsEc2InstanceType,
		WorkerAzureVMSize:        cmdWorkerAzureVMSize,
		WorkerMemorySizeGB:       flags.WorkerMemorySizeGB,
		WorkerNumCPUs:            flags.WorkerNumCPUs,
		WorkersMax:               flags.WorkersMax,
		WorkersMin:               flags.WorkersMin,
		WorkerStorageSizeGB:      flags.WorkerStorageSizeGB,
	}
}

type creationResult struct {
	// cluster ID
	id string
	// location to fetch details on new cluster from
	location string
	// cluster definition assembled, v4 compatible
	definitionV4 *types.ClusterDefinitionV4
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

  cat my-cluster.yaml | gsctl create cluster -f -
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
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().IntVarP(&cmdAvailabilityZones, "availability-zones", "", 0, "Number of availability zones to use on AWS. Default is 1.")
	Command.Flags().StringVarP(&cmdInputYAMLFile, "file", "f", "", "Path to a cluster definition YAML file. Use '-' to read from STDIN.")
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
func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments()

	headline := ""
	subtext := ""

	err := verifyPreconditions(args)
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
func printResult(cmd *cobra.Command, positionalArgs []string) {
	// use arguments as passed from command line via cobra
	args := collectArguments()

	result, err := addCluster(args)
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
			if args.InputYAMLFile != "" {
				subtext = "Please specify an owner organization for the cluster in your definition file or set one via the --owner flag."
			}
		case errors.IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("If you specify workers in your definition file, you'll have to specify at least %d worker nodes for a useful cluster.", limits.MinimumNumWorkers)
		case errors.IsYAMLNotParseable(err):
			headline = "Could not parse YAML"
			if args.InputYAMLFile == "-" {
				subtext = "The YAML data given via STDIN could not be parsed into a cluster definition."
			} else {
				subtext = fmt.Sprintf("The YAML data read from file '%s' could not be parsed into a cluster definition.", args.InputYAMLFile)
			}
		case errors.IsYAMLFileNotReadable(err):
			headline = "Could not read YAML file"
			subtext = fmt.Sprintf("The file '%s' could not be read. Please make sure that it is readable and contains valid YAML.", args.InputYAMLFile)
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
	if !args.DryRun {
		if result.definitionV4.Name != "" {
			fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' is launching.", result.definitionV4.Name, result.id, result.definitionV4.Owner))
		} else {
			fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' is launching.", result.id, result.definitionV4.Owner))
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

// verifyPreconditions checks preconditions and returns an error in case.
func verifyPreconditions(args Arguments) error {
	// logged in?
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	// false flag combination?
	if args.InputYAMLFile != "" {
		if args.NumWorkers != 0 || args.WorkerNumCPUs != 0 || args.WorkerMemorySizeGB != 0 || args.WorkerStorageSizeGB != 0 || args.WorkerAwsEc2InstanceType != "" || args.WorkerAzureVMSize != "" {
			return microerror.Mask(errors.ConflictingFlagsError)
		}
	}

	// validate number of workers specified by flag
	if args.NumWorkers > 0 && (args.WorkersMax > 0 || args.WorkersMin > 0) {
		return microerror.Mask(errors.ConflictingWorkerFlagsUsedError)
	}
	if args.NumWorkers > 0 && args.NumWorkers < limits.MinimumNumWorkers {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.WorkersMax > 0 && args.WorkersMax < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.WorkersMin > 0 && args.WorkersMin < int64(limits.MinimumNumWorkers) {
		return microerror.Mask(errors.NotEnoughWorkerNodesError)
	}
	if args.WorkersMin > 0 && args.WorkersMax > 0 && args.WorkersMin > args.WorkersMax {
		return microerror.Mask(errors.WorkersMinMaxInvalidError)
	}

	// validate number of CPUs specified by flag
	if args.WorkerNumCPUs > 0 && args.WorkerNumCPUs < limits.MinimumWorkerNumCPUs {
		return microerror.Mask(errors.NotEnoughCPUCoresPerWorkerError)
	}

	// validate memory size specified by flag
	if args.WorkerMemorySizeGB > 0 && args.WorkerMemorySizeGB < limits.MinimumWorkerMemorySizeGB {
		return microerror.Mask(errors.NotEnoughMemoryPerWorkerError)
	}

	// validate storage size specified by flag
	if args.WorkerStorageSizeGB > 0 && args.WorkerStorageSizeGB < limits.MinimumWorkerStorageSizeGB {
		return microerror.Mask(errors.NotEnoughStoragePerWorkerError)
	}

	if args.WorkerAwsEc2InstanceType != "" || args.WorkerAzureVMSize != "" {
		// check for incompatibilities
		if args.WorkerNumCPUs != 0 || args.WorkerMemorySizeGB != 0 || args.WorkerStorageSizeGB != 0 {
			return microerror.Mask(errors.IncompatibleSettingsError)
		}
	}

	return nil
}

// addCluster actually adds a cluster, interpreting all the input configuration
// and returning a structured result.
func addCluster(args Arguments) (*creationResult, error) {
	result := &creationResult{
		definitionV4: &types.ClusterDefinitionV4{},
	}

	var err error

	if args.InputYAMLFile == "-" {
		// Definintion coming from STDIN.
		yamlString := ""
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			yamlString += scanner.Text() + "\n"
		}

		if err := scanner.Err(); err != nil {
			return nil, microerror.Mask(err)
		}

		result.definitionV4, err = readDefinitionFromYAMLV4([]byte(yamlString))
		if err != nil {
			return nil, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
		}
	} else if args.InputYAMLFile != "" {
		// definition from file (and optionally flags)
		result.definitionV4, err = readDefinitionFromFileV4(args.FileSystem, args.InputYAMLFile)
		if err != nil {
			return nil, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
		}
	}

	// Let user-provided arguments (flags) overwrite/extend definition from YAML.
	updateDefinitionFromFlagsV4(result.definitionV4, args)

	// Validate definition
	if result.definitionV4.Owner == "" {
		return nil, microerror.Mask(errors.ClusterOwnerMissingError)
	}

	// Validations based on definition file.
	// For validations based on command line flags, see validatePreConditions()
	if args.InputYAMLFile != "" {
		// number of workers
		if len(result.definitionV4.Workers) > 0 && len(result.definitionV4.Workers) < limits.MinimumNumWorkers {
			return nil, microerror.Mask(errors.NotEnoughWorkerNodesError)
		}
	}

	// create JSON API call payload to catch and handle errors early
	addClusterBody := createAddClusterBodyV4(result.definitionV4)
	_, marshalErr := json.Marshal(addClusterBody)
	if marshalErr != nil {
		return nil, microerror.Maskf(errors.CouldNotCreateJSONRequestBodyError, marshalErr.Error())
	}

	// Preview in YAML format
	if args.Verbose {
		fmt.Println("\nDefinition for the requested cluster:")
		d, marshalErr := yaml.Marshal(addClusterBody)
		if marshalErr != nil {
			log.Fatalf("error: %v", marshalErr)
		}
		fmt.Printf(color.CyanString(string(d)))
		fmt.Println()
	}

	if !args.DryRun {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(result.definitionV4.Owner))

		clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// TODO: switch between v4 or v5 cluster creation

		auxParams := clientWrapper.DefaultAuxiliaryParams()
		auxParams.ActivityName = createClusterActivityName
		// perform API call
		response, err := clientWrapper.CreateClusterV4(addClusterBody, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// success
		result.location = response.Location
		result.id = strings.Split(result.location, "/")[3]
	}

	return result, nil
}
