package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

type cpuDefinition struct {
	Cores int `yaml:"cores,omitempty"`
}

type memoryDefinition struct {
	SizeGB int `yaml:"size_gb,omitempty"`
}

type storageDefinition struct {
	SizeGB int `yaml:"size_gb,omitempty"`
}

type nodeDefinition struct {
	Memory  memoryDefinition  `yaml:"memory,omitempty"`
	CPU     cpuDefinition     `yaml:"cpu,omitempty"`
	Storage storageDefinition `yaml:"storage,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

type clusterDefinition struct {
	Name              string           `yaml:"name,omitempty"`
	Owner             string           `yaml:"owner,omitempty"`
	KubernetesVersion string           `yaml:"kubernetes_version,omitempty"`
	Workers           []nodeDefinition `yaml:"workers,omitempty"`
}

const (
	// TODO: These defaults should come from the API
	minimumNumWorkers          int = 1
	minimumWorkerNumCPUs       int = 1
	minimumWorkerMemorySizeGB  int = 1
	minimumWorkerStorageSizeGB int = 1

	createClusterActivityName string = "create-cluster"
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
		PreRunE: checkAddCluster,
		Run:     addCluster,
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
	// number of CPUs per worker as required via flag on execution
	cmdWorkerNumCPUs int
	// RAM size in GB per worker as required via flag on execution
	cmdWorkerMemorySizeGB int
	// Local storage in GB per worker as required via flag on execution
	cmdWorkerStorageSizeGB int
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
	CreateClusterCommand.Flags().IntVarP(&cmdWorkerMemorySizeGB, "memory-gb", "", 0, "RAM per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().IntVarP(&cmdWorkerStorageSizeGB, "storage-gb", "", 0, "Local storage size per worker node. Can't be used with -f|--file.")
	CreateClusterCommand.Flags().BoolVarP(&cmdDryRun, "dry-run", "", false, "If set, the cluster won't be created. Useful with -v|--verbose.")

	CreateCommand.AddCommand(CreateClusterCommand)
}

// checks preconditions
func checkAddCluster(cmd *cobra.Command, args []string) error {
	// logged in?
	if config.Config.Token == "" && cmdToken == "" {
		s := color.RedString("You are not logged in.\n\n")
		return errors.New(s + "Use '" + config.ProgramName + " login' to login or '--auth-token' to pass a valid auth token.")
	}
	// false flag combination?
	if cmdInputYAMLFile != "" {
		if cmdNumWorkers != 0 || cmdWorkerNumCPUs != 0 || cmdWorkerMemorySizeGB != 0 || cmdWorkerStorageSizeGB != 0 {
			s := color.RedString("Conflicting flags used\n\n")
			return errors.New(s + "When specifying a definition via a YAML file, certain flags must not be used.\n")
		}
	} else {
		if cmdNumWorkers == 0 && (cmdWorkerNumCPUs != 0 || cmdWorkerMemorySizeGB != 0 || cmdWorkerStorageSizeGB != 0) {
			s := color.RedString("Number of worker nodes required\n\n")
			return errors.New(s + "When requiring specific worker node specification details, you must also specify the number of worker nodes.\n")
		}
	}

	// validate number of workers specified by flag
	if cmdNumWorkers > 0 && cmdNumWorkers < minimumNumWorkers {
		s := color.RedString("\nNot enough worker nodes specified\n")
		s = s + "You'll need at least " + string(minimumNumWorkers) + " worker nodes for a useful cluster.\n\n"
		return errors.New(s)
	}

	// validate number of CPUs specified by flag
	if cmdWorkerNumCPUs > 0 && cmdWorkerNumCPUs < minimumWorkerNumCPUs {
		s := color.RedString("\nNot enough CPUs per worker specified\n")
		s = s + "You'll need at least " + string(minimumWorkerNumCPUs) + " CPU cores per worker node.\n\n"
		return errors.New(s)
	}

	// validate memory size specified by flag
	if cmdWorkerMemorySizeGB > 0 && cmdWorkerMemorySizeGB < minimumWorkerMemorySizeGB {
		s := color.RedString("\nNot enough Memory per worker specified\n")
		s = s + "You'll need at least " + string(minimumWorkerMemorySizeGB) + " GB per worker node.\n\n"
		return errors.New(s)
	}

	// validate storage size specified by flag
	if cmdWorkerStorageSizeGB > 0 && cmdWorkerStorageSizeGB < minimumWorkerStorageSizeGB {
		s := color.RedString("\nNot enough Memory per worker specified\n")
		s = s + "You'll need at least " + string(minimumWorkerStorageSizeGB) + " GB per worker node.\n\n"
		return errors.New(s)
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
func enhanceDefinitionWithFlags(def *clusterDefinition) {
	if cmdClusterName != "" {
		def.Name = cmdClusterName
	}
	if cmdKubernetesVersion != "" {
		def.KubernetesVersion = cmdKubernetesVersion
	}
	if cmdOwner != "" {
		def.Owner = cmdOwner
	}
}

// createDefinitionFromFlags creates a clusterDefinition based on the
// flags the user has given
func createDefinitionFromFlags() clusterDefinition {
	def := clusterDefinition{}
	if cmdClusterName != "" {
		def.Name = cmdClusterName
	}
	if cmdKubernetesVersion != "" {
		def.KubernetesVersion = cmdKubernetesVersion
	}
	if cmdOwner != "" {
		def.Owner = cmdOwner
	}
	if cmdNumWorkers != 0 {
		workers := []nodeDefinition{}
		for i := 0; i < cmdNumWorkers; i++ {
			worker := nodeDefinition{}
			if cmdWorkerNumCPUs != 0 {
				worker.CPU = cpuDefinition{Cores: cmdWorkerNumCPUs}
			}
			if cmdWorkerStorageSizeGB != 0 {
				worker.Storage = storageDefinition{SizeGB: cmdWorkerStorageSizeGB}
			}
			if cmdWorkerMemorySizeGB != 0 {
				worker.Memory = memoryDefinition{SizeGB: cmdWorkerMemorySizeGB}
			}
			workers = append(workers, worker)
			//fmt.Printf("%#v\n", worker)
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
		ndmWorker.Memory = gsclientgen.V4NodeDefinitionMemory{SizeGb: int32(dWorker.Memory.SizeGB)}
		ndmWorker.Cpu = gsclientgen.V4NodeDefinitionCpu{Cores: int32(dWorker.CPU.Cores)}
		ndmWorker.Storage = gsclientgen.V4NodeDefinitionStorage{SizeGb: int32(dWorker.Storage.SizeGB)}
		ndmWorker.Labels = dWorker.Labels
		a.Workers = append(a.Workers, ndmWorker)
	}

	return a
}

// interprets arguments/flags, shows validation results, eventually submits create request
func addCluster(cmd *cobra.Command, args []string) {
	var definition clusterDefinition
	var err error
	if cmdInputYAMLFile != "" {
		// definition from file (and optionally flags)
		definition, err = readDefinitionFromFile(cmdInputYAMLFile)
		if err != nil {
			fmt.Println(color.RedString("Could not read file '%s'", cmdInputYAMLFile))
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		enhanceDefinitionWithFlags(&definition)
	} else {
		// definition from flags only
		definition = createDefinitionFromFlags()
	}

	// Validate result and give feedback
	if definition.Owner == "" {
		// try using default organization
		if config.Config.Organization != "" {
			definition.Owner = config.Config.Organization
		} else {
			fmt.Printf("\n%s\n", color.RedString("No owner organization set"))
			if cmdInputYAMLFile != "" {
				fmt.Println("Please specify an owner organization for the cluster in your definition file or set one via the --owner flag.\n")
			} else {
				fmt.Println("Please specify an owner organization for the cluster via the --owner flag.\n")
			}
			os.Exit(1)
		}
	}

	// Validations based on definition file.
	// For validations based on command line flags, see checkAddCluster()
	if cmdInputYAMLFile != "" {
		// number of workers
		if len(definition.Workers) > 0 && len(definition.Workers) < minimumNumWorkers {
			fmt.Printf("\n%s\n", color.RedString("Not enough worker nodes specified"))
			fmt.Printf("If you specify workers in your definition file, you'll have to specify at least %d worker nodes for a useful cluster.\n\n", minimumNumWorkers)
			os.Exit(1)
		}
	}

	// Preview in YAML format
	if cmdVerbose {
		fmt.Println("\nDefinition for the requested cluster:")
		d, err := yaml.Marshal(definition)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf(color.CyanString(string(d)))
		fmt.Println()
	}

	// create JSON API call payload to catch and handle errors early
	addClusterBody := createAddClusterBody(definition)
	if cmdVerbose {
		_, err := json.Marshal(addClusterBody)
		if err != nil {
			fmt.Println()
			fmt.Println(color.RedString("Could not create JSON body for API request"))
			fmt.Printf("Error message: %s\n\n", err)
			fmt.Println("Please contact Giant Swarm via support@giantswarm.io in case you need any help.")
			os.Exit(1)
		}
	}

	if !cmdDryRun {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(definition.Owner))

		// perform API call
		authHeader := "giantswarm " + config.Config.Token
		if cmdToken != "" {
			// command line flag overwrites
			authHeader = "giantswarm " + cmdToken
		}
		client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
		responseBody, apiResponse, _ := client.AddCluster(authHeader, addClusterBody, requestIDHeader, createClusterActivityName, cmdLine)

		// handle API result
		if responseBody.Code == "RESOURCE_CREATED" {
			clusterID := strings.Split(apiResponse.Header["Location"][0], "/")[3]
			fmt.Println(color.GreenString("New cluster with ID '%s' is launching.", clusterID))
			fmt.Println("Add key-pair and settings to kubectl using\n")
			fmt.Printf("    %s\n\n", color.YellowString("gsctl create kubeconfig --cluster="+clusterID))
		} else {
			fmt.Println()
			fmt.Println(color.RedString("Could not create cluster"))
			fmt.Printf("Error message: %s\n", responseBody.Message)
			fmt.Printf("Error code: %s\n", responseBody.Code)
			fmt.Println(fmt.Sprintf("Raw response body:\n%v", string(apiResponse.Payload)))
			fmt.Println("Please contact Giant Swarm via support@giantswarm.io in case you need any help.")
			os.Exit(1)
		}
	}

}
