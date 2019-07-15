// Package cluster implements the 'show cluster' command.
package cluster

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
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
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	// Time after which a new cluster should be up, roughly.
	clusterCreationExpectedDuration = 20 * time.Minute
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
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return showClusterArguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
		clusterID:   "",
		verbose:     flags.CmdVerbose,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowClusterArguments()
	err := verifyShowClusterPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyShowClusterPreconditions(args showClusterArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	return nil
}

// getClusterDetails returns details for one cluster.
func getClusterDetails(clusterID, activityName string) (*models.V4ClusterDetailsResponse, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientV2.GetCluster(clusterID, auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			switch clientErr.HTTPStatusCode {
			case http.StatusForbidden:
				return nil, microerror.Mask(errors.AccessForbiddenError)
			case http.StatusUnauthorized:
				return nil, microerror.Mask(errors.NotAuthorizedError)
			case http.StatusNotFound:
				return nil, microerror.Mask(errors.ClusterNotFoundError)
			case http.StatusInternalServerError:
				return nil, microerror.Mask(errors.InternalServerError)
			}
		}

		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}

func getOrgCredentials(orgName, credentialID, activityName string) (*models.V4GetCredentialResponse, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientV2.GetCredential(orgName, credentialID, auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			switch clientErr.HTTPStatusCode {
			case http.StatusForbidden:
				return nil, microerror.Mask(errors.AccessForbiddenError)
			case http.StatusUnauthorized:
				return nil, microerror.Mask(errors.NotAuthorizedError)
			case http.StatusNotFound:
				return nil, microerror.Mask(errors.CredentialNotFoundError)
			case http.StatusInternalServerError:
				return nil, microerror.Mask(errors.InternalServerError)
			}
		}

		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}

// sumWorkerCPUs adds up the worker's CPU cores
func sumWorkerCPUs(numWorkers int, workerDetails []*models.V4ClusterDetailsResponseWorkersItems) uint {
	sum := numWorkers * int(workerDetails[0].CPU.Cores)
	return uint(sum)
}

// sumWorkerStorage adds up the worker's storage
func sumWorkerStorage(numWorkers int, workerDetails []*models.V4ClusterDetailsResponseWorkersItems) float64 {
	sum := float64(numWorkers) * workerDetails[0].Storage.SizeGb
	return sum
}

// sumWorkerMemory adds up the worker's memory
func sumWorkerMemory(numWorkers int, workerDetails []*models.V4ClusterDetailsResponseWorkersItems) float64 {
	sum := float64(numWorkers) * workerDetails[0].Memory.SizeGb
	return sum
}

// printResult fetches cluster info from the API, which involves
// several API calls, and prints the output.
func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowClusterArguments()
	args.clusterID = cmdLineArgs[0]

	if args.verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster %s", args.clusterID))
	}

	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	clusterDetailsChan := make(chan *models.V4ClusterDetailsResponse)
	clusterDetailsErrChan := make(chan error)

	go func(chan *models.V4ClusterDetailsResponse, chan error) {
		clusterDetails, err := getClusterDetails(args.clusterID, showClusterActivityName)
		clusterDetailsChan <- clusterDetails
		clusterDetailsErrChan <- err
	}(clusterDetailsChan, clusterDetailsErrChan)

	clusterStatusChan := make(chan *client.ClusterStatus)
	clusterStatusErrChan := make(chan error)

	go func(chan *client.ClusterStatus, chan error) {
		auxParams := clientV2.DefaultAuxiliaryParams()
		auxParams.ActivityName = showClusterActivityName

		status, err := clientV2.GetClusterStatus(args.clusterID, auxParams)
		clusterStatusChan <- status
		clusterStatusErrChan <- err
	}(clusterStatusChan, clusterStatusErrChan)

	clusterDetails := <-clusterDetailsChan
	clusterDetailsErr := <-clusterDetailsErrChan
	clusterStatus := <-clusterStatusChan
	clusterStatusErr := <-clusterStatusErrChan

	if clusterDetailsErr != nil {
		errors.HandleCommonErrors(clusterDetailsErr)
		client.HandleErrors(clusterDetailsErr)

		var headline = ""
		var subtext = ""

		switch {
		case errors.IsClusterNotFoundError(clusterDetailsErr):
			headline = "Cluster not found"
			subtext = "The cluster with this ID could not be found. Please use 'gsctl list clusters' to list all available clusters."
		case errors.IsCredentialNotFoundError(clusterDetailsErr):
			headline = "Credential not found"
			subtext = "Credentials with the given ID could not be found."
		case clusterDetailsErr.Error() == "":
			return
		default:
			headline = clusterDetailsErr.Error()
		}

		// Print error output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	var credentialDetails *models.V4GetCredentialResponse
	//var credentialDetailsErr error
	if clusterDetailsErr == nil && clusterDetails.CredentialID != "" {
		if args.verbose {
			fmt.Println(color.WhiteString("Fetching details for credential %s", clusterDetails.CredentialID))
		}

		credentialDetails, _ = getOrgCredentials(clusterDetails.Owner, clusterDetails.CredentialID, showClusterActivityName)
	}

	// Calculate worker node count.
	numWorkers := 0
	if clusterStatus != nil && clusterStatus.Cluster.Nodes != nil {
		// Count all nodes as workers which are not explicitly marked as master.
		for _, node := range clusterStatus.Cluster.Nodes {
			val, ok := node.Labels["role"]
			if !ok {
				// Workaround for k8s 1.14 because the label changed.
				val, ok = node.Labels["kubernetes.io/role"]
			}
			if ok && val == "master" {
				// don't count this
			} else {
				numWorkers++
			}
		}
	}

	// print table
	output := []string{}

	created := util.ParseDate(clusterDetails.CreateDate)

	output = append(output, color.YellowString("ID:")+"|"+clusterDetails.ID)

	if clusterDetails.Name != "" {
		output = append(output, color.YellowString("Name:")+"|"+clusterDetails.Name)
	} else {
		output = append(output, color.YellowString("Name:")+"|n/a")
	}
	output = append(output, color.YellowString("Created:")+"|"+util.ShortDate(created))
	output = append(output, color.YellowString("Organization:")+"|"+clusterDetails.Owner)

	if credentialDetails != nil && credentialDetails.ID != "" {
		if credentialDetails.Aws != nil {
			parts := strings.Split(credentialDetails.Aws.Roles.Awsoperator, ":")
			if len(parts) > 3 {
				output = append(output, color.YellowString("AWS account:")+"|"+parts[4])
			} else {
				output = append(output, color.YellowString("AWS account:")+"|n/a")
			}
		} else if credentialDetails.Azure != nil {
			output = append(output, color.YellowString("Azure subscription:")+"|"+credentialDetails.Azure.Credential.SubscriptionID)
			output = append(output, color.YellowString("Azure tenant:")+"|"+credentialDetails.Azure.Credential.TenantID)
		}
	}

	output = append(output, color.YellowString("Kubernetes API endpoint:")+"|"+clusterDetails.APIEndpoint)

	if len(clusterDetails.AvailabilityZones) > 0 {
		sort.Strings(clusterDetails.AvailabilityZones)
		output = append(output, color.YellowString("Availability Zones:")+"|"+strings.Join(clusterDetails.AvailabilityZones, ", "))
	}

	if clusterDetails.ReleaseVersion != "" {
		output = append(output, color.YellowString("Release version:")+"|"+clusterDetails.ReleaseVersion)
	} else {
		output = append(output, color.YellowString("Release version:")+"|n/a")
	}

	// Instance type / VM size
	if clusterDetails.Workers[0].Aws != nil && clusterDetails.Workers[0].Aws.InstanceType != "" {
		output = append(output, color.YellowString("Worker EC2 instance type:")+"|"+clusterDetails.Workers[0].Aws.InstanceType)
	} else if clusterDetails.Workers[0].Azure != nil && clusterDetails.Workers[0].Azure.VMSize != "" {
		output = append(output, color.YellowString("Worker VM size:")+"|"+clusterDetails.Workers[0].Azure.VMSize)
	}

	// scaling info
	scalingInfo := "n/a"
	if clusterDetails.Scaling != nil {
		if clusterDetails.Scaling.Min == clusterDetails.Scaling.Max {
			scalingInfo = fmt.Sprintf("pinned at %d", clusterDetails.Scaling.Min)
		} else {
			scalingInfo = fmt.Sprintf("autoscaling between %d and %d", clusterDetails.Scaling.Min, clusterDetails.Scaling.Max)
		}
	}
	output = append(output, color.YellowString("Worker node scaling:")+"|"+scalingInfo)

	// what the autoscaler tries to reach as a target (only interesting if not pinned)
	if clusterStatus != nil && clusterStatus.Cluster != nil && clusterDetails.Scaling != nil && clusterDetails.Scaling.Min != clusterDetails.Scaling.Max {
		output = append(output, color.YellowString("Desired worker node count:")+"|"+fmt.Sprintf("%d", clusterStatus.Cluster.Scaling.DesiredCapacity))
	}

	// current number of workers
	output = append(output, color.YellowString("Worker nodes running:")+"|"+fmt.Sprintf("%d", numWorkers))

	output = append(output, color.YellowString("CPU cores in workers:")+"|"+fmt.Sprintf("%d", sumWorkerCPUs(numWorkers, clusterDetails.Workers)))
	output = append(output, color.YellowString("RAM in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerMemory(numWorkers, clusterDetails.Workers)))

	if clusterDetails.Kvm != nil {
		output = append(output, color.YellowString("Storage in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerStorage(numWorkers, clusterDetails.Workers)))
	}

	if clusterDetails.Kvm != nil && len(clusterDetails.Kvm.PortMappings) > 0 {
		for _, portMapping := range clusterDetails.Kvm.PortMappings {
			output = append(output, color.YellowString(fmt.Sprintf("Ingress port for %s:", portMapping.Protocol))+"|"+fmt.Sprintf("%d", portMapping.Port))
		}
	}

	fmt.Println(columnize.SimpleFormat(output))

	// Here we inform about problems fetching the cluster status, which happens regularly for new clusters
	// and isn't a problem.
	if clusterStatusErr != nil {
		fmt.Println("\nInfo: Could not fetch cluster status, so the worker node count displayed might derive from the actual number.")

		if time.Since(created) < clusterCreationExpectedDuration {
			fmt.Println("This is expected for clusters which are most likely still in creation.")
		}
	}
}
