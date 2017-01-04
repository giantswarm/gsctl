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
	SizeGB int `yaml:"size_gb"`
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
	Masters           []nodeDefinition `yaml:"masters,omitempty"`
	Workers           []nodeDefinition `yaml:"workers,omitempty"`
}

const (
	// TODO: These defaults should come from the API
	defaultNumWorkers          int = 3
	defaultNumMasters          int = 1
	defaultWorkerNumCPUs       int = 1
	defaultWorkerMemorySizeGB  int = 2
	defaultWorkerStorageSizeGB int = 10

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

Alternatively, if detailed specification of individual worker nodes and/or the
master node is required, a YAML file can be passed using the --file|-f flag. In
this case, some command line flags like --name and --owner can be used to
extend/overwrite the definition given as a file.

Examples:

  gsctl create cluster --owner=myorg --num-workers=5 --num-cpus=2 --memory-gb=8 --storage-gb=100

  gsctl create cluster --file my-cluster.yaml

	gsctl create cluster --num-workers=2 --dry-run --verbose`,
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
	CreateClusterCommand.Flags().BoolVarP(&cmdDryRun, "dry-run", "", false, "If set, the cluster won't be created. Useful with -v|--versbose.")

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
	return nil
}

// readDefinitionFromFile reads a cluster definition from a YAML config file
func readDefinitionFromFile(filePath string) (clusterDefinition, error) {
	myDef := clusterDefinition{}
	data, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		return myDef, readErr
	}

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

// creates a gsclientgen.AddClusterBodyModel from clusterDefinition
func createAddClusterBody(d clusterDefinition) gsclientgen.AddClusterBodyModel {
	a := gsclientgen.AddClusterBodyModel{}
	a.Name = d.Name
	a.Owner = d.Owner
	a.KubernetesVersion = d.KubernetesVersion

	for _, dWorker := range d.Workers {
		ndmWorker := gsclientgen.NodeDefinitionModel{}
		ndmWorker.Memory = gsclientgen.NodeDefinitionModelMemory{SizeGb: int32(dWorker.Memory.SizeGB)}
		ndmWorker.Cpu = gsclientgen.NodeDefinitionModelCpu{Cores: int32(dWorker.CPU.Cores)}
		ndmWorker.Storage = gsclientgen.NodeDefinitionModelStorage{SizeGb: int32(dWorker.Storage.SizeGB)}
		ndmWorker.Labels = dWorker.Labels
		a.Workers = append(a.Workers, ndmWorker)
	}

	for _, dMaster := range d.Masters {
		ndmMaster := gsclientgen.NodeDefinitionModel{}
		ndmMaster.Memory = gsclientgen.NodeDefinitionModelMemory{SizeGb: int32(dMaster.Memory.SizeGB)}
		ndmMaster.Cpu = gsclientgen.NodeDefinitionModelCpu{Cores: int32(dMaster.CPU.Cores)}
		ndmMaster.Storage = gsclientgen.NodeDefinitionModelStorage{SizeGb: int32(dMaster.Storage.SizeGB)}
		a.Masters = append(a.Masters, ndmMaster)
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
			fmt.Println(color.RedString("No owner organization set"))
			if cmdInputYAMLFile != "" {
				fmt.Println("Please specify an owner organization for the cluster in your definition file or set one via the --owner flag.")
			} else {
				fmt.Println("Please specify an owner organization for the cluster via the --owner flag.")
			}
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
		client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
		responseBody, apiResponse, _ := client.AddCluster(authHeader, addClusterBody, requestIDHeader, createClusterActivityName, cmdLine)

		// handle API result
		if responseBody.Code == 201 {
			clusterID := strings.Split(apiResponse.Header["Location"][0], "/")[3]
			fmt.Printf("New cluster with ID '%s' is launching. You can go ahead and add kubectl credentials using\n\n", color.CyanString(clusterID))
			fmt.Printf("    %s\n\n", color.YellowString("gsctl create kubeconfig -c "+clusterID))
		} else {
			fmt.Println()
			fmt.Println(color.RedString("Could not create cluster"))
			fmt.Printf("Error message: %s\n", responseBody.Message)
			fmt.Printf("Response code: %d\n", responseBody.Code)
			fmt.Printf("Request ID: %s\n\n", requestIDHeader)
			fmt.Println("Please contact Giant Swarm via support@giantswarm.io in case you need any help.")
			os.Exit(1)
		}
	}

}
