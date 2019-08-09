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
	"github.com/giantswarm/gsctl/nodespec"
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

// Arguments specifies all the arguments to be used for our business function.
type Arguments struct {
	apiEndpoint       string
	authToken         string
	scheme            string
	clusterID         string
	userProvidedToken string
	verbose           bool
}

// collectArguments fills arguments from user input, config, and environment.
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		scheme:            scheme,
		clusterID:         "",
		userProvidedToken: flags.CmdToken,
		verbose:           flags.CmdVerbose,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()
	err := verifyShowClusterPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyShowClusterPreconditions(args Arguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	return nil
}

// getClusterDetailsV4 returns details for one cluster.
func getClusterDetailsV4(args Arguments) (*models.V4ClusterDetailsResponse, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetClusterV4(args.clusterID, auxParams)
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
func getClusterDetailsV5(args Arguments) (*models.V5ClusterDetailsResponse, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetClusterV5(args.clusterID, auxParams)
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

func getOrgCredentials(orgName, credentialID string, args Arguments) (*models.V4GetCredentialResponse, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.GetCredential(orgName, credentialID, auxParams)
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
func getClusterDetails(args Arguments) (
	*models.V4ClusterDetailsResponse,
	*models.V5ClusterDetailsResponse,
	*models.V5GetNodePoolsResponse,
	*client.ClusterStatus,
	*models.V4GetCredentialResponse,
	error) {

	var clusterDetailsV4 *models.V4ClusterDetailsResponse
	var clusterDetailsV5 *models.V5ClusterDetailsResponse
	var clusterStatus *client.ClusterStatus
	var nodePools *models.V5GetNodePoolsResponse

	// first try v5
	if args.verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster via v5 API endpoint."))
	}
	clusterDetailsV5, v5Err := getClusterDetailsV5(args)
	if v5Err == nil {
		// fetch node pools
		clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
		if err != nil {
			return nil, nil, nil, nil, nil, microerror.Mask(err)
		}

		// perform API call
		auxParams := clientWrapper.DefaultAuxiliaryParams()
		auxParams.ActivityName = activityName
		response, err := clientWrapper.GetNodePools(args.clusterID, auxParams)
		if err != nil {
			return nil, nil, nil, nil, nil, microerror.Mask(err)
		}
		nodePools = &response.Payload

	} else {
		// If this is a 404 error, we assume the cluster is not a V5 one.
		// If this is a "Malformed response" error, we assume the API is not capable of
		// handling V5 yet. TODO: This can be phased out once the API is up-to-date.
		// In both these case we continue below, otherwise we return the error.
		if !errors.IsClusterNotFoundError(v5Err) && !clienterror.IsMalformedResponseError(v5Err) {
			return nil, nil, nil, nil, nil, microerror.Mask(v5Err)
		}

		// Fall back to v4.
		if args.verbose {
			fmt.Println(color.WhiteString("No usable v5 response. Fetching details for cluster via v4 API endpoint."))
		}

		var clusterDetailsV4Err error
		clusterDetailsV4, clusterDetailsV4Err = getClusterDetailsV4(args)
		if clusterDetailsV4Err != nil {
			// At this point, every error is a sign of something unexpected, so
			// simply return.
			return nil, nil, nil, nil, nil, microerror.Mask(clusterDetailsV4Err)
		}

		clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
		if err != nil {
			return nil, nil, nil, nil, nil, microerror.Mask(err)
		}

		if args.verbose {
			fmt.Println(color.WhiteString("Fetching status for v4 cluster."))
		}
		auxParams := clientWrapper.DefaultAuxiliaryParams()
		auxParams.ActivityName = activityName
		var clusterStatusErr error
		clusterStatus, clusterStatusErr = clientWrapper.GetClusterStatus(args.clusterID, auxParams)
		if clusterStatusErr != nil {
			// Return an error if it is something else than 404 Not Found,
			// as 404s are expected during cluster creation.
			if !errors.IsClusterNotFoundError(clusterStatusErr) {
				return nil, nil, nil, nil, nil, microerror.Mask(clusterStatusErr)
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
			credentialDetails, credentialDetailsErr = getOrgCredentials(clusterOwner, credentialID, args)
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

	return clusterDetailsV4, clusterDetailsV5, nodePools, clusterStatus, credentialDetails, nil
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
	args := collectArguments()
	args.clusterID = cmdLineArgs[0]

	if args.verbose {
		fmt.Println(color.WhiteString("Fetching details for cluster %s.", args.clusterID))
	}

	clusterDetailsV4, clusterDetailsV5, nodePools, clusterStatus, credentialDetails, err := getClusterDetails(args)
	if err != nil {
		errors.HandleCommonErrors(err)
	}

	if clusterDetailsV4 != nil {
		printV4Result(args, clusterDetailsV4, clusterStatus, credentialDetails)
	} else if clusterDetailsV5 != nil {
		printV5Result(args, clusterDetailsV5, credentialDetails, nodePools)
	}
}

// printV4Result prints the detils for a V4 cluster.
func printV4Result(args Arguments, clusterDetails *models.V4ClusterDetailsResponse, clusterStatus *client.ClusterStatus, credentialDetails *models.V4GetCredentialResponse) {
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

	output = append(output, color.YellowString("ID:")+"|"+clusterDetails.ID)
	output = append(output, color.YellowString("Name:")+"|"+stringOrPlaceholder(clusterDetails.Name))
	output = append(output, color.YellowString("Created:")+"|"+formatDate(clusterDetails.CreateDate))
	output = append(output, color.YellowString("Organization:")+"|"+clusterDetails.Owner)
	output = append(output, color.YellowString("Kubernetes API endpoint:")+"|"+clusterDetails.APIEndpoint)
	output = append(output, color.YellowString("Release version:")+"|"+stringOrPlaceholder(clusterDetails.ReleaseVersion))

	// BYOC credentials.
	if credentialDetails != nil && credentialDetails.ID != "" {
		output = append(output, formatCredentialDetails(credentialDetails)...)
	}

	if len(clusterDetails.AvailabilityZones) > 0 {
		sort.Strings(clusterDetails.AvailabilityZones)
		output = append(output, color.YellowString("Availability Zones:")+"|"+strings.Join(clusterDetails.AvailabilityZones, ", "))
	}

	// Instance type / VM size
	if clusterDetails.Workers[0].Aws != nil && clusterDetails.Workers[0].Aws.InstanceType != "" {
		output = append(output, color.YellowString("Worker EC2 instance type:")+"|"+clusterDetails.Workers[0].Aws.InstanceType)
	} else if clusterDetails.Workers[0].Azure != nil && clusterDetails.Workers[0].Azure.VMSize != "" {
		output = append(output, color.YellowString("Worker VM size:")+"|"+clusterDetails.Workers[0].Azure.VMSize)
	}

	// scaling info
	scalingInfo := ""
	if clusterDetails.Scaling != nil {
		if clusterDetails.Scaling.Min == clusterDetails.Scaling.Max {
			scalingInfo = fmt.Sprintf("pinned at %d", clusterDetails.Scaling.Min)
		} else {
			scalingInfo = fmt.Sprintf("autoscaling between %d and %d", clusterDetails.Scaling.Min, clusterDetails.Scaling.Max)
		}
	}
	output = append(output, color.YellowString("Worker node scaling:")+"|"+stringOrPlaceholder(scalingInfo))

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

	// KVM ingress port mappings
	if clusterDetails.Kvm != nil && len(clusterDetails.Kvm.PortMappings) > 0 {
		for _, portMapping := range clusterDetails.Kvm.PortMappings {
			output = append(output, color.YellowString(fmt.Sprintf("Ingress port for %s:", portMapping.Protocol))+"|"+fmt.Sprintf("%d", portMapping.Port))
		}
	}

	fmt.Println(columnize.SimpleFormat(output))
}

// printV5Result prints details for a v5 clsuter.
func printV5Result(args Arguments, details *models.V5ClusterDetailsResponse,
	credentialDetails *models.V4GetCredentialResponse,
	nodePools *models.V5GetNodePoolsResponse) {

	// clusterTable is the table for cluster information.
	clusterTable := []string{}

	clusterTable = append(clusterTable, color.YellowString("Cluster ID:")+"|"+details.ID)
	clusterTable = append(clusterTable, color.YellowString("Name:")+"|"+stringOrPlaceholder(details.Name))
	clusterTable = append(clusterTable, color.YellowString("Created:")+"|"+formatDate(details.CreateDate))
	clusterTable = append(clusterTable, color.YellowString("Organization:")+"|"+details.Owner)
	clusterTable = append(clusterTable, color.YellowString("Kubernetes API endpoint:")+"|"+details.APIEndpoint)
	clusterTable = append(clusterTable, color.YellowString("Master availability zone:")+"|"+details.Master.AvailabilityZone)
	clusterTable = append(clusterTable, color.YellowString("Release version:")+"|"+details.ReleaseVersion)

	// BYOC credentials.
	if credentialDetails != nil && credentialDetails.ID != "" {
		clusterTable = append(clusterTable, formatCredentialDetails(credentialDetails)...)
	}

	// TODO: Add KVM ingress port mappings here
	// once KVM is supported in V5.

	// Aggregate of node pools.
	if nodePools != nil && len(*nodePools) > 0 {
		clusterTable = append(clusterTable, formatNodePoolDetails(nodePools)...)
	}

	fmt.Println(columnize.SimpleFormat(clusterTable))
}

// formatDate takes a date/time string from the API and returns a formated version.
func formatDate(dt string) string {
	created := util.ParseDate(dt)
	return util.ShortDate(created)
}

// stringOrPlaceholder takes an input string and returns either the string or,
// if string is empty, or the "n/a" placeholder.
func stringOrPlaceholder(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}

// formatCredentialDetails returns the info table rows erquired to print details about
// the credential given.
func formatCredentialDetails(credentialDetails *models.V4GetCredentialResponse) []string {
	rows := []string{}

	if credentialDetails.Aws != nil {
		parts := strings.Split(credentialDetails.Aws.Roles.Awsoperator, ":")
		if len(parts) > 3 {
			rows = append(rows, color.YellowString("AWS account:")+"|"+parts[4])
		} else {
			rows = append(rows, color.YellowString("AWS account:")+"|n/a")
		}
	} else if credentialDetails.Azure != nil {
		rows = append(rows, color.YellowString("Azure subscription:")+"|"+credentialDetails.Azure.Credential.SubscriptionID)
		rows = append(rows, color.YellowString("Azure tenant:")+"|"+credentialDetails.Azure.Credential.TenantID)
	}

	return rows
}

func formatNodePoolDetails(nodePools *models.V5GetNodePoolsResponse) []string {
	rows := []string{}

	cpus := 0
	ramGB := 0
	numNodes := 0
	numNodePools := len(*nodePools)

	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		fmt.Println(color.RedString("Error: Cannot provide info on AWS instance types. Details: %s", err))
	}

	if nodePools != nil && numNodePools > 0 {
		for _, np := range *nodePools {
			numNodes += int(np.Status.NodesReady)

			// Provider: AWS
			if np.NodeSpec.Aws != nil && np.NodeSpec.Aws.InstanceType != "" {
				it, err := awsInfo.GetInstanceTypeDetails(np.NodeSpec.Aws.InstanceType)
				if err != nil {
					fmt.Println(color.YellowString("Warning: Cannot provide info on AWS instance type '%s'. Please kindly report this to the Giant Swarm support team.", np.NodeSpec.Aws.InstanceType))
				} else {
					cpus += it.CPUCores * int(np.Status.NodesReady)
					ramGB += it.MemorySizeGB * int(np.Status.NodesReady)
				}
			}
		}
	}

	rows = append(rows, color.YellowString("Size:")+fmt.Sprintf("|%d nodes in %d node pools", numNodes, numNodePools))
	rows = append(rows, color.YellowString("CPUs in nodes:")+fmt.Sprintf("|%d", cpus))
	rows = append(rows, color.YellowString("RAM in nodes (GB):")+fmt.Sprintf("|%d", ramGB))

	return rows
}
