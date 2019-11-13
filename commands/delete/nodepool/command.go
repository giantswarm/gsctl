// Package nodepool implements the "delete nodepool" command.
package nodepool

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command is the cobra command for 'gsctl delete nodepool'
	Command = &cobra.Command{
		Use:     "nodepool <cluster-id>/<nodepool-id>",
		Aliases: []string{"np"},
		// Args: cobra.ExactArgs(1) guarantees that cobra will fail if no positional argument is given.
		Args:  cobra.ExactArgs(1),
		Short: "Delete a node pool",
		Long: `Delete new node pool from a cluster.

This command allows to delete a node pool.

Deleting a node pool means that all worker nodes in the pool will be drained,
cordoned and then terminated.

In case you are running workloads on the node pool you want to delete,
make sure that there is at least one other node pool with capacity to
schedule the workloads. Also check whether label selectors, taints and
tolerations will allow scheduling on other pool's worker nodes. The best
way to observe this is by manually cordoning and draining the pool's
worker nodes and checking workload's node assignments, before issuing
the 'delete nodepool' command.

Note: Data stored outside of persistent volumes will be lost and there is
no way to undo this.

Examples:

  To delete node pool 'np1id' from cluster 'f01r4', use this command:

    gsctl delete nodepool f01r4/np1id

  To prevent the confirmation questions, apply --force:

    gsctl delete nodepool f01r4/np1id --force
`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}
)

const (
	activityName = "delete-nodepool"
)

func init() {
	initFlags()
}

// initFlags initializes flags in a re-usable way, so we can call it from multiple tests.
func initFlags() {
	Command.ResetFlags()
	Command.Flags().BoolVarP(&flags.Force, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
}

// Arguments defines the arguments this command can take into consideration.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterID         string
	Force             bool
	NodePoolID        string
	UserProvidedToken string
	Verbose           bool
}

// collectArguments populates an arguments struct with values both from command flags,
// from config, and potentially from built-in defaults.
func collectArguments(positionalArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	parts := strings.Split(positionalArgs[0], "/")

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterID:         parts[0],
		Force:             flags.Force,
		NodePoolID:        parts[1],
		UserProvidedToken: flags.Token,
		Verbose:           flags.Verbose,
	}
}

func verifyPreconditions(args Arguments) error {
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.ClusterID == "" {
		return microerror.Mask(errors.ClusterIDMissingError)
	}
	if args.NodePoolID == "" {
		return microerror.Mask(errors.NodePoolIDMissingError)
	}

	return nil
}

func printValidation(cmd *cobra.Command, positionalArgs []string) {
	var err error
	args := collectArguments(positionalArgs)
	if err == nil {
		err = verifyPreconditions(args)
	}

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

// deleteNodePool is the business function sending our deletion request to the API
// and returning true for success or an error.
func deleteNodePool(args Arguments) (bool, error) {
	// confirmation
	if !args.Force {
		question := fmt.Sprintf("Do you really want to delete node pool '%s' from cluster '%s'?", args.NodePoolID, args.ClusterID)
		confirmed := confirm.Ask(question)
		if !confirmed {
			return false, nil
		}
	}

	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return false, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	_, err = clientWrapper.DeleteNodePool(args.ClusterID, args.NodePoolID, auxParams)
	if clienterror.IsAccessForbiddenError(err) {
		return false, microerror.Mask(errors.AccessForbiddenError)
	} else if clienterror.IsNotFoundError(err) {
		// Check whether the cluster exists
		_, detailsErr := clientWrapper.GetClusterV5(args.ClusterID, auxParams)
		if detailsErr == nil {
			// Cluster exists, node pool does not exist.
			return false, microerror.Mask(errors.NodePoolNotFoundError)
		}

		_, detailsErr = clientWrapper.GetClusterV4(args.ClusterID, auxParams)
		if detailsErr == nil {
			// Cluster exists, but is v4, so cannot have node pools.
			// TODO: use errors.ClusterDoesNotSupportNodePoolsError when available
		}

		return false, microerror.Mask(errors.ClusterNotFoundError)
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func printResult(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments(positionalArgs)

	deleted, err := deleteNodePool(args)
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

	if deleted {
		fmt.Println(color.GreenString("Node pool '%s' in cluster '%s' will be deleted as soon as all workloads are terminated.", args.NodePoolID, args.ClusterID))
	} else if args.Verbose {
		fmt.Println(color.WhiteString("Aborted."))
	}
}
