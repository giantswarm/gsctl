// Package nodepool implements the "show nodepool" command.
package nodepool

import (
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

var (
	// Command is the cobra command for 'gsctl create nodepool'
	Command = &cobra.Command{
		Hidden:  false, // TODO: set to true before merging into master!
		Use:     "nodepool <cluster-id>",
		Aliases: []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Create a node pool",
		Long: `Add a new node pool to a cluster.

This command allows to create a new node pool within a cluster. Node pools
are groups of wortker nodes sharing a common configuration. Create different
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
	cmdName                 string
	cmdNumAvailabilityZones int
	cmdAvailabilityZones    []string
)

const (
	activityName = "create-nodepool"
)

func init() {
	Command.Flags().StringVarP(&cmdName, "name", "n", "", "name or purpose description of the node pool")
	Command.Flags().IntVarP(&cmdNumAvailabilityZones, "num-availability-zones", "", 0, "Number of availability zones to use. Default is 1.")
	Command.Flags().StringSliceVarP(&cmdAvailabilityZones, "availability-zones", "", []string{}, "List of availability zones to use, instead of setting a number. Use comma to separate values.")
	Command.Flags().StringVarP(&cmdAwsEc2InstanceType, "aws-instance-type", "", "", "EC2 instance type to use for workers, e. g. 'm5.2xlarge'")
	// TODO: size min/max flags
}

type arguments struct {
	apiEndpoint           string
	authToken             string
	availabilityZonesList []string
	availabilityZonesNum  int
	clusterID             string
	instanceType          string
	name                  string
	scheme                string
	verbose               bool
}

func defaultArguments(positionalArgs []string) arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return arguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
		clusterID:   positionalArgs[0],
		verbose:     flags.CmdVerbose,
		name:        cmdName,
	}
}

func verifyPreconditions(args arguments, positionalArgs []string) error {
	parsedArgs := defaultArguments(positionalArgs)
	if config.Config.Token == "" && parsedArgs.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	if parsedArgs.clusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}

	if len(parsedArgs.availabilityZonesList) > 0 && parsedArgs.availabilityZonesNum > 0 {
		return microerror.Maskf(errors.ConflictingFlagsError, "the flags --availability-zones and --num-availability-zones cannot be combined.")
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {

}

func printResult(cmd *cobra.Command, positionalArgs []string) {

}
