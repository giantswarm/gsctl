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
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/config"
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
	activityName = "show-cluster"
)

type showClusterArguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
	clusterID   string
	verbose     bool
}

func defaultArguments() showClusterArguments {
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
	args := defaultArguments()
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

// getClusterDetailsV4 returns details for one cluster.
func getClusterDetailsV4(clusterID string) (*models.V4ClusterDetailsResponse, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientV2.GetClusterV4(clusterID, auxParams)
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

// getClusterDetailsV5 returns details for one cluster, supporting node pools.
func getClusterDetailsV5(clusterID string) (*models.V5ClusterDetailsResponse, error) {
	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientV2.GetClusterV5(clusterID, auxParams)
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

func getOrgCredentials(orgName, credentialID string) (*models.V4GetCredentialResponse, error) {
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

// getClusterDetails returns all cluster details that are of interest
// in the context of this command:
//
// - cluster details (v4 or v5)
// - cluster status (v4 only)
// - credential details, in case this is a BYOC cluster
func getClusterDetails(args showClusterArguments) (
	*models.V4ClusterDetailsResponse,
	*models.V5ClusterDetailsResponse,
	*client.ClusterStatus,
	*models.V4GetCredentialResponse,
	error) {

	var clusterDetailsV4 *models.V4ClusterDetailsResponse
	var clusterDetailsV5 *models.V5ClusterDetailsResponse
	var clusterStatus *client.ClusterStatus

	// first try v5
	if args.verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster via v5 API endpoint"))
	}
	clusterDetailsV5, v5Err := getClusterDetailsV5(args.clusterID)
	if v5Err != nil {
		// If this is a 404 error, continue with v4 fallback below.
		// Otherwise return error.
		if !errors.IsClusterNotFoundError(v5Err) {
			return nil, nil, nil, nil, microerror.Mask(v5Err)
		}

		// Fall back to v4.
		if args.verbose {
			fmt.Println(color.WhiteString("Fetching details for cluster via v4 API endpoint"))
		}

		var clusterDetailsV4Err error
		clusterDetailsV4, clusterDetailsV4Err = getClusterDetailsV4(args.clusterID)
		if clusterDetailsV4Err != nil {
			// At this point, every error is a sign of something unexpected, so
			// simply return.
			return nil, nil, nil, nil, microerror.Mask(clusterDetailsV4Err)
		}

		clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
		if err != nil {
			return nil, nil, nil, nil, microerror.Mask(err)
		}

		if args.verbose {
			fmt.Println(color.WhiteString("Fetching status for v4 cluster"))
		}
		auxParams := clientV2.DefaultAuxiliaryParams()
		auxParams.ActivityName = activityName
		var clusterStatusErr error
		clusterStatus, clusterStatusErr = clientV2.GetClusterStatus(args.clusterID, auxParams)
		if clusterStatusErr != nil {
			// Return an error if it is something else than 404 Not Found,
			// as 404s are expected during cluster creation.
			if !errors.IsClusterNotFoundError(clusterStatusErr) {
				return nil, nil, nil, nil, microerror.Mask(clusterStatusErr)
			}
		}
	}

	var credentialDetails *models.V4GetCredentialResponse
	{
		credentialID := ""
		clusterOwner := ""
		var created time.Time

		if clusterDetailsV4 != nil {
			credentialID = clusterDetailsV4.CredentialID
			clusterOwner = clusterDetailsV4.Owner
			created = util.ParseDate(clusterDetailsV4.CreateDate)
		} else if clusterDetailsV5 != nil {
			credentialID = clusterDetailsV5.CredentialID
			clusterOwner = clusterDetailsV5.Owner
			created = util.ParseDate(clusterDetailsV5.CreateDate)
		}

		if credentialID != "" {
			if args.verbose {
				fmt.Println(color.WhiteString("Fetching credential details for organization %s", clusterOwner))
			}

			var credentialDetailsErr error
			credentialDetails, credentialDetailsErr = getOrgCredentials(clusterOwner, credentialID)
			if credentialDetailsErr != nil {
				if time.Since(created) < clusterCreationExpectedDuration {
					fmt.Println("This is expected for clusters which are most likely still in creation.")
				}
				// Print any error occurring here, but don't return, as this is non-critical.
				fmt.Printf(color.YellowString("Warning: credential details for org %s (credential ID %s) could not be fetched.\n", clusterOwner, credentialID))
				fmt.Printf("Error details: %s\n", credentialDetailsErr)
			}
		}
	}

	return clusterDetailsV4, clusterDetailsV5, clusterStatus, credentialDetails, nil
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
	args := defaultArguments()
	args.clusterID = cmdLineArgs[0]

	if args.verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster %s", args.clusterID))
	}

	clusterDetailsV4, _, clusterStatus, credentialDetails, err := getClusterDetails(args)
	if err != nil {
		errors.HandleCommonErrors(err)
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

	created := util.ParseDate(clusterDetailsV4.CreateDate)

	output = append(output, color.YellowString("ID:")+"|"+clusterDetailsV4.ID)

	if clusterDetailsV4.Name != "" {
		output = append(output, color.YellowString("Name:")+"|"+clusterDetailsV4.Name)
	} else {
		output = append(output, color.YellowString("Name:")+"|n/a")
	}
	output = append(output, color.YellowString("Created:")+"|"+util.ShortDate(created))
	output = append(output, color.YellowString("Organization:")+"|"+clusterDetailsV4.Owner)

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

	output = append(output, color.YellowString("Kubernetes API endpoint:")+"|"+clusterDetailsV4.APIEndpoint)

	if len(clusterDetailsV4.AvailabilityZones) > 0 {
		sort.Strings(clusterDetailsV4.AvailabilityZones)
		output = append(output, color.YellowString("Availability Zones:")+"|"+strings.Join(clusterDetailsV4.AvailabilityZones, ", "))
	}

	if clusterDetailsV4.ReleaseVersion != "" {
		output = append(output, color.YellowString("Release version:")+"|"+clusterDetailsV4.ReleaseVersion)
	} else {
		output = append(output, color.YellowString("Release version:")+"|n/a")
	}

	// Instance type / VM size
	if clusterDetailsV4.Workers[0].Aws != nil && clusterDetailsV4.Workers[0].Aws.InstanceType != "" {
		output = append(output, color.YellowString("Worker EC2 instance type:")+"|"+clusterDetailsV4.Workers[0].Aws.InstanceType)
	} else if clusterDetailsV4.Workers[0].Azure != nil && clusterDetailsV4.Workers[0].Azure.VMSize != "" {
		output = append(output, color.YellowString("Worker VM size:")+"|"+clusterDetailsV4.Workers[0].Azure.VMSize)
	}

	// scaling info
	scalingInfo := "n/a"
	if clusterDetailsV4.Scaling != nil {
		if clusterDetailsV4.Scaling.Min == clusterDetailsV4.Scaling.Max {
			scalingInfo = fmt.Sprintf("pinned at %d", clusterDetailsV4.Scaling.Min)
		} else {
			scalingInfo = fmt.Sprintf("autoscaling between %d and %d", clusterDetailsV4.Scaling.Min, clusterDetailsV4.Scaling.Max)
		}
	}
	output = append(output, color.YellowString("Worker node scaling:")+"|"+scalingInfo)

	// what the autoscaler tries to reach as a target (only interesting if not pinned)
	if clusterStatus != nil && clusterStatus.Cluster != nil && clusterDetailsV4.Scaling != nil && clusterDetailsV4.Scaling.Min != clusterDetailsV4.Scaling.Max {
		output = append(output, color.YellowString("Desired worker node count:")+"|"+fmt.Sprintf("%d", clusterStatus.Cluster.Scaling.DesiredCapacity))
	}

	// current number of workers
	output = append(output, color.YellowString("Worker nodes running:")+"|"+fmt.Sprintf("%d", numWorkers))

	output = append(output, color.YellowString("CPU cores in workers:")+"|"+fmt.Sprintf("%d", sumWorkerCPUs(numWorkers, clusterDetailsV4.Workers)))
	output = append(output, color.YellowString("RAM in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerMemory(numWorkers, clusterDetailsV4.Workers)))

	if clusterDetailsV4.Kvm != nil {
		output = append(output, color.YellowString("Storage in worker nodes (GB):")+"|"+fmt.Sprintf("%.2f", sumWorkerStorage(numWorkers, clusterDetailsV4.Workers)))
	}

	if clusterDetailsV4.Kvm != nil && len(clusterDetailsV4.Kvm.PortMappings) > 0 {
		for _, portMapping := range clusterDetailsV4.Kvm.PortMappings {
			output = append(output, color.YellowString(fmt.Sprintf("Ingress port for %s:", portMapping.Protocol))+"|"+fmt.Sprintf("%d", portMapping.Port))
		}
	}

	fmt.Println(columnize.SimpleFormat(output))
}
