// Package nodepool implements the "create nodepool" command.
package nodepool

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
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
		Use:     "nodepool <cluster-id>",
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

- Name: will be "Unnamed node pool <n>".
- Availability zones: the node pool will use 1 zone selected randomly.
- Instance type: the default instance type of the installation will be
  used. Check 'gsctl info' to find out what that is.
- Scaling settings: the minimum and maximum size will be set to 3,
  meaning that autoscaling is disabled.

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

    gsctl create nodepool f01r4 --num-availability-zones 2

  To set one or several specific zones to use, give a list of zone names
  or letters.

    gsctl create nodepool f01r4 --availability-zones b,c,d

  Here is how you specify the instance type to use:

    gsctl create nodepool f01r4 --aws-instance-type m4.2xlarge

  The initial node pool size is set by adjusting the lower and upper
  size limit like this:

    gsctl create nodepool f01r4 --nodes-min 3 --nodes-max 10
`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	cmdAwsEc2InstanceType   string
	cmdAvailabilityZonesNum int
	cmdAvailabilityZones    []string
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
}

// Arguments defines the arguments this command can take into consideration.
type Arguments struct {
	APIEndpoint           string
	AuthToken             string
	AvailabilityZonesList []string
	AvailabilityZonesNum  int
	ClusterID             string
	InstanceType          string
	Name                  string
	ScalingMax            int64
	ScalingMin            int64
	Scheme                string
	UserProvidedToken     string
	Verbose               bool
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
		zones, err = expandZones(zones, endpoint, flags.Token)
		if err != nil {
			return Arguments{}, microerror.Mask(err)
		}
	}

	if flags.WorkersMin > 0 && flags.WorkersMax == 0 {
		flags.WorkersMax = flags.WorkersMin
	} else if flags.WorkersMax > 0 && flags.WorkersMin == 0 {
		flags.WorkersMin = flags.WorkersMax
	}

	return Arguments{
		APIEndpoint:           endpoint,
		AuthToken:             token,
		AvailabilityZonesList: zones,
		AvailabilityZonesNum:  cmdAvailabilityZonesNum,
		ClusterID:             positionalArgs[0],
		InstanceType:          flags.WorkerAwsEc2InstanceType,
		Name:                  flags.Name,
		ScalingMax:            flags.WorkersMax,
		ScalingMin:            flags.WorkersMin,
		Scheme:                scheme,
		UserProvidedToken:     flags.Token,
		Verbose:               flags.Verbose,
	}, nil
}

// expandZones takes a list of alphabetical letters and returns a list of
// availability zones. Example:
//
// ["a", "b"] -> ["eu-central-1a", "eu-central-1b"]
//
func expandZones(zones []string, endpoint, userProvidedToken string) ([]string, error) {
	clientWrapper, err := client.NewWithConfig(endpoint, userProvidedToken)
	if err != nil {
		return []string{}, microerror.Mask(err)
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
	if config.Config.Token == "" && args.AuthToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	if args.ClusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
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

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	var err error
	args, err := collectArguments(positionalArgs)
	if err == nil {
		err = verifyPreconditions(args)
	}

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
	client.HandleErrors(err)

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
func createNodePool(args Arguments) (*result, error) {
	r := &result{}

	requestBody := &models.V5AddNodePoolRequest{
		Name: args.Name,
	}
	if args.InstanceType != "" {
		requestBody.NodeSpec = &models.V5AddNodePoolRequestNodeSpec{
			Aws: &models.V5AddNodePoolRequestNodeSpecAws{
				InstanceType: args.InstanceType,
			},
		}
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
		requestBody.Scaling = &models.V5AddNodePoolRequestScaling{
			Min: args.ScalingMin,
			Max: args.ScalingMax,
		}
	}

	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return r, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.CreateNodePool(args.ClusterID, requestBody, auxParams)

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
	var r *result

	args, err := collectArguments(positionalArgs)
	if err == nil {
		r, err = createNodePool(args)
	}

	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

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
		fmt.Println("The API call to create a node pooll apparently has been successful, however")
		fmt.Println("no useful response has been returned. Please report this problem to the")
		fmt.Println("Giant Swarm support team. Thank you!")
		os.Exit(1)
	}

	fmt.Println(color.GreenString("New node pool '%s' (ID '%s') in cluster '%s' is launching.", r.nodePoolName, r.nodePoolID, args.ClusterID))
	fmt.Printf("Use this command to inspect details for the new node pool:\n\n")
	fmt.Println(color.YellowString("    gsctl show nodepool %s/%s", args.ClusterID, r.nodePoolID))
}
