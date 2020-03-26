// Package nodepool implements the "create nodepool" command.
package nodepool

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the cobra command for 'gsctl create nodepool'
	Command = &cobra.Command{
		Use:     "nodepool <cluster-name/cluster-id>",
		Aliases: []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Create a node pool",
		Long: `Add a new node pool to a cluster.

This command allows to create a new node pool within a cluster. Node pools
are groups of worker nodes sharing a common configuration. Create different
node pools to serve workloads with different resource requirements, different
availability zone spreading etc. Node pools are also scaled independently.

Note that some attributes of node pools cannot be changed later. These are:

- ID - will be generated on creation
- Availability zone assignment
- Instance type used

When creating a node pool, all arguments (except for the cluster ID) are
optional. Where an argument is not given, a default will be applied as
follows:

- Name: will be "Unnamed node pool".
- Availability zones: the node pool will use 1 zone selected randomly.
- Instance type: the default instance type of the installation will be
  used. Check 'gsctl info' to find out what that is.
- Scaling settings: the minimum will be 3 and maximum 10 nodes.

Examples:

  The simplest invocation creates a node pool with default attributes
  in the cluster specified.

    gsctl create nodepool f01r4

  We recommend to set a descriptive name, to tell the node pool apart
  from others.

    gsctl create nodepool f01r4  --name "Batch jobs"

  Assigning the node pool to availabilty zones can be done in several
  ways. If you only want to ensure that several zones are used, specify
  a number liker like this:

    gsctl create nodepool "Cluster name" --num-availability-zones 2

  To set one or several specific zones to use, give a list of zone names
  or letters.

    gsctl create nodepool f01r4 --availability-zones b,c,d

  Here is how you specify the instance type to use:

    gsctl create nodepool "Cluster name" --aws-instance-type m4.2xlarge

  The initial node pool size is set by adjusting the lower and upper
  size limit like this:

    gsctl create nodepool f01r4 --nodes-min 3 --nodes-max 10

  To use 50% spot instances in a node pool and making sure to always have
  three on-demand instances you can create your node pool like this:

    gsctl create nodepool f01r4 --nodes-min 3 --nodes-max 10 \
	  --aws-on-demand-base-capacity 3 \
	  --aws-on-demand-percentage-above-base-capacity 50

  To use similar instances in your node pool to the one that you defined
  you can create your node pool like this (the list is maintained by
  Giant Swarm for now eg. if you select m5.xlarge the node pool can fall
  back on m4.xlarge too):

    gsctl create nodepool f01r4 --aws-instance-type m4.xlarge \
	  --aws-use-alike-instance-types

`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	cmdAwsEc2InstanceType   string
	cmdAvailabilityZonesNum int
	cmdAvailabilityZones    []string

	arguments Arguments
)

const (
	activityName = "create-nodepool"
)

func init() {
	initFlags()
}

// initFlags initializes flags in a re-usable way, so we can call it from multiple tests.
func initFlags() {
	Command.ResetFlags()
	Command.Flags().StringVarP(&flags.Name, "name", "n", "", "name or purpose description of the node pool")
	Command.Flags().IntVarP(&cmdAvailabilityZonesNum, "num-availability-zones", "", 0, "Number of availability zones to use. Default is 1.")
	Command.Flags().StringSliceVarP(&cmdAvailabilityZones, "availability-zones", "", nil, "List of availability zones to use, instead of setting a number. Use comma to separate values.")
	Command.Flags().StringVarP(&flags.WorkerAwsEc2InstanceType, "aws-instance-type", "", "", "EC2 instance type to use for workers, e. g. 'm5.2xlarge'")
	Command.Flags().Int64VarP(&flags.WorkersMin, "nodes-min", "", 0, "Minimum number of worker nodes for the node pool.")
	Command.Flags().Int64VarP(&flags.WorkersMax, "nodes-max", "", 0, "Maximum number of worker nodes for the node pool.")
	Command.Flags().BoolVarP(&flags.AWSUseAlikeInstanceTypes, "aws-use-alike-instance-types", "", false, "Use similar instance type in your node pool. This list is maintained by Giant Swarm at the moment. Eg if you select m5.xlarge then the node pool can fall back on m4.xlarge too.")
	Command.Flags().Int64VarP(&flags.AWSOnDemandBaseCapacity, "aws-on-demand-base-capacity", "", 0, "Number of on-demand instances that this node pool needs to have until spot instances are used. Default is 0")
	Command.Flags().Int64VarP(&flags.AWSOnDemandPercentageAboveBaseCapacity, "aws-on-demand-percentage-above-base-capacity", "", 100, "Percentage of on-demand instances used once the on-demand base capacity is fullfilled. A number of 40 would mean that 60 percent will be spot instances. Default is 100.")
}

// Arguments defines the arguments this command can take into consideration.
type Arguments struct {
	APIEndpoint                         string
	AuthToken                           string
	AvailabilityZonesList               []string
	AvailabilityZonesNum                int
	ClusterNameOrID                     string
	InstanceType                        string
	UseAlikeInstanceTypes               bool
	OnDemandBaseCapacity                int64
	OnDemandPercentageAboveBaseCapacity int64
	Name                                string
	ScalingMax                          int64
	ScalingMin                          int64
	Scheme                              string
	UserProvidedToken                   string
	Verbose                             bool
}

type result struct {
	nodePoolID            string
	nodePoolName          string
	availabilityZonesList []string
}

// collectArguments populates an arguments struct with values both from command flags,
// from config, and potentially from built-in defaults.
func collectArguments(positionalArgs []string) (Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	var err error

	zones := cmdAvailabilityZones
	if zones != nil && len(zones) > 0 {
		zones, err = expandZones(zones, endpoint, flags.Token, flags.Verbose)
		if err != nil {
			return Arguments{}, microerror.Mask(err)
		}
	}

	fmt.Println(flags.AWSOnDemandPercentageAboveBaseCapacity)

	return Arguments{
		APIEndpoint:                         endpoint,
		AuthToken:                           token,
		AvailabilityZonesList:               zones,
		AvailabilityZonesNum:                cmdAvailabilityZonesNum,
		ClusterNameOrID:                     positionalArgs[0],
		InstanceType:                        flags.WorkerAwsEc2InstanceType,
		UseAlikeInstanceTypes:               flags.AWSUseAlikeInstanceTypes,
		OnDemandBaseCapacity:                flags.AWSOnDemandBaseCapacity,
		OnDemandPercentageAboveBaseCapacity: flags.AWSOnDemandPercentageAboveBaseCapacity,
		Name:                                flags.Name,
		ScalingMax:                          flags.WorkersMax,
		ScalingMin:                          flags.WorkersMin,
		Scheme:                              scheme,
		UserProvidedToken:                   flags.Token,
		Verbose:                             flags.Verbose,
	}, nil
}

// expandZones takes a list of alphabetical letters and returns a list of
// availability zones. Example:
//
// ["a", "b"] -> ["eu-central-1a", "eu-central-1b"]
//
func expandZones(zones []string, endpoint, userProvidedToken string, verbose bool) ([]string, error) {
	clientWrapper, err := client.NewWithConfig(endpoint, userProvidedToken)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	if verbose {
		fmt.Println(color.WhiteString("Fetching installation info to validate availability zones"))
	}

	info, err := clientWrapper.GetInfo(nil)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	out := []string{}
	for _, letter := range zones {
		if len(letter) == 1 {
			letter = info.Payload.General.Datacenter + strings.ToLower(letter)
		}
		out = append(out, letter)
	}

	return out, nil
}

func verifyPreconditions(args Arguments) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.AuthToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	if args.ClusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	}

	// AZ flags plausibility
	if len(args.AvailabilityZonesList) > 0 && args.AvailabilityZonesNum > 0 {
		return microerror.Maskf(errors.ConflictingFlagsError, "the flags --availability-zones and --num-availability-zones cannot be combined.")
	}

	// Scaling flags plausibility
	if args.ScalingMin > 0 && args.ScalingMax > 0 {
		if args.ScalingMin > args.ScalingMax {
			return microerror.Mask(errors.WorkersMinMaxInvalidError)
		}
	}

	// OnDemandPercentageAboveBaseCapacity check percentage
	if args.OnDemandPercentageAboveBaseCapacity < 0 || args.OnDemandPercentageAboveBaseCapacity > 100 {
		return microerror.Mask(errors.NotPercentage)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	var err error

	arguments, err = collectArguments(positionalArgs)
	if err == nil {
		err = verifyPreconditions(arguments)
	}

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	case errors.IsConflictingFlagsError(err):
		headline = "Conflicting flags used"
		subtext = "The flags --availability-zones and --num-availability-zones must not be used together."
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

// createNodePool is the business function sending our creation request to the API
// and returning either a proper result or an error.
func createNodePool(args Arguments, clusterID string, clientWrapper *client.Wrapper) (*result, error) {
	r := &result{}

	requestBody := &models.V5AddNodePoolRequest{
		Name: args.Name,
	}

	requestBody.NodeSpec = &models.V5AddNodePoolRequestNodeSpec{
		Aws: &models.V5AddNodePoolRequestNodeSpecAws{
			UseAlikeInstanceTypes: args.UseAlikeInstanceTypes,
			InstanceDistribution: &models.V5AddNodePoolRequestNodeSpecAwsInstanceDistribution{
				OnDemandBaseCapacity:                &args.OnDemandBaseCapacity,
				OnDemandPercentageAboveBaseCapacity: &args.OnDemandPercentageAboveBaseCapacity,
			},
		},
	}
	if args.InstanceType != "" {
		requestBody.NodeSpec.Aws.InstanceType = args.InstanceType
	}
	if args.AvailabilityZonesList != nil && len(args.AvailabilityZonesList) > 0 {
		requestBody.AvailabilityZones = &models.V5AddNodePoolRequestAvailabilityZones{
			Zones: args.AvailabilityZonesList,
		}
	} else if args.AvailabilityZonesNum != 0 {
		requestBody.AvailabilityZones = &models.V5AddNodePoolRequestAvailabilityZones{
			Number: int64(args.AvailabilityZonesNum),
		}
	}
	if args.ScalingMin != 0 || args.ScalingMax != 0 {
		requestBody.Scaling = &models.V5AddNodePoolRequestScaling{}
		if args.ScalingMin != 0 {
			requestBody.Scaling.Min = args.ScalingMin
		}
		if args.ScalingMax != 0 {
			requestBody.Scaling.Max = args.ScalingMax
		}
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	if args.Verbose {
		fmt.Println(color.WhiteString("Submitting node pool creation request"))
		bodyJSON, _ := json.Marshal(requestBody)
		fmt.Println(color.WhiteString("Request body: ") + string(bodyJSON))
	}

	response, err := clientWrapper.CreateNodePool(clusterID, requestBody, auxParams)

	if err != nil {
		// return specific error types for the cases we care about most.
		if clienterror.IsAccessForbiddenError(err) {
			return nil, microerror.Mask(errors.AccessForbiddenError)
		}
		if clienterror.IsNotFoundError(err) {
			return nil, microerror.Mask(errors.ClusterNotFoundError)
		}
		if clienterror.IsBadRequestError(err) {
			return nil, microerror.Maskf(errors.BadRequestError, err.Error())
		}
		if clienterror.IsInternalServerError(err) {
			return nil, microerror.Maskf(errors.InternalServerError, err.Error())
		}

		return r, microerror.Mask(err)
	}

	r.nodePoolID = response.Payload.ID
	r.nodePoolName = response.Payload.Name
	r.availabilityZonesList = response.Payload.AvailabilityZones

	return r, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	clientWrapper, err := client.NewWithConfig(arguments.APIEndpoint, arguments.UserProvidedToken)
	if err != nil {
		err = microerror.Mask(err)
	}
	clusterID, err := clustercache.GetID(arguments.APIEndpoint, arguments.ClusterNameOrID, clientWrapper)
	if err != nil {
		err = microerror.Mask(err)
	}

	r, err := createNodePool(arguments, clusterID, clientWrapper)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		headline := ""
		subtext := ""

		switch {
		// If there are specific errors to handle, add them here.
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

	if r == nil {
		// This is unlikely, but hey.
		fmt.Println(color.RedString("No response returned"))
		fmt.Println("The API call to create a node pool apparently has been successful, however")
		fmt.Println("no useful response has been returned. Please report this problem to the")
		fmt.Println("Giant Swarm support team. Thank you!")
		os.Exit(1)
	}

	fmt.Println(color.GreenString("New node pool '%s' (ID '%s') in cluster '%s' is launching.", r.nodePoolName, r.nodePoolID, clusterID))
	fmt.Printf("Use this command to inspect details for the new node pool:\n\n")
	fmt.Println(color.YellowString("    gsctl show nodepool %s/%s", clusterID, r.nodePoolID))
	fmt.Printf("\n")
}
