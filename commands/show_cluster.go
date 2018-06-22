package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (

	// ShowClusterCommand performs the "show cluster" function
	ShowClusterCommand = &cobra.Command{
		Use:   "cluster",
		Short: "Show cluster details",
		Long: `Display details of a cluster

Examples:

  gsctl show cluster c7t2o
`,

		// PreRun checks a few general things, like authentication.
		PreRun: showClusterPreRunOutput,

		// Run calls the business function and prints results and errors.
		Run: showClusterRunOutput,
	}
)

const (
	showClusterActivityName = "show-cluster"
)

type showClusterArguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
	clusterID   string
	verbose     bool
}

func defaultShowClusterArguments() showClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdScheme)

	return showClusterArguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
		clusterID:   "",
		verbose:     cmdVerbose,
	}
}

func init() {
	ShowCommand.AddCommand(ShowClusterCommand)
}

func showClusterPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowClusterArguments()
	err := verifyShowClusterPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyShowClusterPreconditions(args showClusterArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(clusterIDMissingError)
	}
	return nil
}

// getClusterDetails returns details for one cluster.
func getClusterDetails(clusterID, scheme, token, endpoint string) (gsclientgen.V4ClusterDetailsModel, error) {
	result := gsclientgen.V4ClusterDetailsModel{}

	// perform API call
	authHeader := scheme + " " + token
	clientConfig := client.Configuration{
		Endpoint:  endpoint,
		UserAgent: config.UserAgent(),
	}

	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(couldNotCreateClientError)
	}

	clusterDetails, apiResp, err := apiClient.GetCluster(authHeader, clusterID,
		requestIDHeader, showClusterActivityName, cmdLine)

	if err != nil {
		if apiResp == nil || apiResp.Response == nil {
			return result, microerror.Mask(noResponseError)
		}

		if apiResp.StatusCode == http.StatusForbidden {
			return result, microerror.Mask(accessForbiddenError)
		}

		return result, microerror.Mask(err)
	}

	switch apiResp.StatusCode {
	case http.StatusUnauthorized:
		return result, microerror.Mask(notAuthorizedError)
	case http.StatusNotFound:
		return result, microerror.Mask(clusterNotFoundError)
	case http.StatusInternalServerError:
		return result, microerror.Mask(internalServerError)
	}

	return *clusterDetails, nil
}

// sumWorkerCPUs adds up the worker's CPU cores
func sumWorkerCPUs(workerDetails []gsclientgen.V4NodeDefinitionResponse) int32 {
	sum := int32(0)
	for _, item := range workerDetails {
		sum = sum + item.Cpu.Cores
	}
	return sum
}

// sumWorkerStorage adds up the worker's storage
func sumWorkerStorage(workerDetails []gsclientgen.V4NodeDefinitionResponse) float32 {
	sum := float32(0.0)
	for _, item := range workerDetails {
		sum = sum + item.Storage.SizeGb
	}
	return sum
}

// sumWorkerMemory adds up the worker's memory
func sumWorkerMemory(workerDetails []gsclientgen.V4NodeDefinitionResponse) float32 {
	sum := float32(0.0)
	for _, item := range workerDetails {
		sum = sum + item.Memory.SizeGb
	}
	return sum
}

func showClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowClusterArguments()
	args.clusterID = cmdLineArgs[0]

	clusterDetails, err := getClusterDetails(args.clusterID, args.scheme,
		args.authToken, args.apiEndpoint)

	if err != nil {
		var headline = ""
		var subtext = ""

		switch {
		case err.Error() == "":
			return
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

	// print table
	output := []string{}

	created := util.ParseDate(clusterDetails.CreateDate)

	output = append(output, color.YellowString("ID:")+"|"+clusterDetails.Id)

	if clusterDetails.Name != "" {
		output = append(output, color.YellowString("Name:")+"|"+clusterDetails.Name)
	} else {
		output = append(output, color.YellowString("Name:")+"|n/a")
	}
	output = append(output, color.YellowString("Created:")+"|"+util.ShortDate(created))
	output = append(output, color.YellowString("Organization:")+"|"+clusterDetails.Owner)
	output = append(output, color.YellowString("Kubernetes API endpoint:")+"|"+clusterDetails.ApiEndpoint)

	if clusterDetails.ReleaseVersion != "" {
		output = append(output, color.YellowString("Release version:")+"|"+clusterDetails.ReleaseVersion)
	} else {
		output = append(output, color.YellowString("Release version:")+"|n/a")
	}

	output = append(output, color.YellowString("Workers:")+"|"+fmt.Sprintf("%d", len(clusterDetails.Workers)))

	// This assumes all nodes use the same instance type.
	if len(clusterDetails.Workers) > 0 {
		if clusterDetails.Workers[0].Aws.InstanceType != "" {
			output = append(output, color.YellowString("Worker instance type:")+"|"+clusterDetails.Workers[0].Aws.InstanceType)
		}

		if clusterDetails.Workers[0].Azure.VmSize != "" {
			output = append(output, color.YellowString("Worker VM size:")+"|"+clusterDetails.Workers[0].Azure.VmSize)
		}
	}

	output = append(output, color.YellowString("CPU cores in workers:")+"|"+fmt.Sprintf("%d", sumWorkerCPUs(clusterDetails.Workers)))
	output = append(output, color.YellowString("RAM in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerMemory(clusterDetails.Workers)))
	output = append(output, color.YellowString("Storage in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerStorage(clusterDetails.Workers)))

	if len(clusterDetails.Kvm.PortMappings) > 0 {
		for _, portMapping := range clusterDetails.Kvm.PortMappings {
			output = append(output, color.YellowString(fmt.Sprintf("Ingress port for %s:", portMapping.Protocol))+"|"+fmt.Sprintf("%d", portMapping.Port))
		}
	}

	fmt.Println(columnize.SimpleFormat(output))
}
