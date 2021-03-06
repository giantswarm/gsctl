package cluster

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/clustercache"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/util"
)

const (
	// upgradeClusterActivityName assigns API requests to named activities
	upgradeClusterActivityName = "upgrade-cluster"

	upgradeDocsURL = "https://docs.giantswarm.io/general/cluster-upgrades/"
)

var (
	// Command performs the "upgrade cluster" function
	Command = &cobra.Command{
		Use:   "cluster",
		Short: "Upgrades a cluster to a newer release version",
		Long: fmt.Sprintf(`Upgrades a cluster to a newer release version.

Upgrades mean the stepwise replacement of the workers, the master nodes and other
building blocks of a cluster with newer versions. See details at

    ` + upgradeDocsURL + `

A cluster will always be upgraded to the subsequent release. To find out what
release version is used currently, use

    gsctl show cluster <cluster-name/cluster-id>

To find out what is the subsequent version, list all available versions using

    gsctl list releases

When in doubt, please contact the Giant Swarm support team before upgrading.

Example:
  gsctl upgrade cluster 6iec4
  gsctl upgrade cluster "Cluster name"
  gsctl upgrade cluster "Cluster name" --release "13.0.0"
`),

		// We use PreRun for general input validation, authentication etc.
		// If something is bad/missing, that function has to exit with a
		// non-zero exit code.
		PreRun: upgradeClusterValidationOutput,

		// Run is the function that actually executes what we want to do.
		Run: upgradeClusterExecutionOutput,
	}

	arguments Arguments
)

// Arguments is the struct to pass to our business function and
// to the validation function.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterNameOrID   string
	Force             bool
	Release           string
	UserProvidedToken string
	Verbose           bool
}

// function to create arguments based on command line flags and config
func collectArguments(positionalArgs []string) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	clusterID := ""
	if len(positionalArgs) > 0 {
		clusterID = positionalArgs[0]
	}

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterNameOrID:   clusterID,
		Force:             flags.Force,
		Release:           flags.Release,
		UserProvidedToken: flags.Token,
		Verbose:           flags.Verbose,
	}
}

type upgradeClusterResult struct {
	clusterID     string
	versionBefore string
	versionAfter  string
}

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().BoolVarP(&flags.Force, "force", "", false, "If set, no interactive confirmation will be required (risky!).")
	Command.Flags().StringVarP(&flags.Release, "release", "", "", "The target release version for the upgrade. If no version is specified, the first version following the running one is selected..")
}

// Prints results of our pre-validation
func upgradeClusterValidationOutput(cmd *cobra.Command, cmdLineArgs []string) {
	arguments = collectArguments(cmdLineArgs)

	headline := ""
	subtext := ""

	err := validateUpgradeClusterPreconditions(arguments, cmdLineArgs)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		switch {
		case err.Error() == "":
			return
		case errors.IsNotLoggedInError(err):
			headline = "You are not logged in."
			subtext = fmt.Sprintf("Use '%s login' to login or '--auth-token' to pass a valid auth token.", config.ProgramName)
		case errors.IsClusterNameOrIDMissingError(err):
			headline = "No cluster name or ID specified."
			subtext = "Please specify which cluster to upgrade by using the cluster name or ID as an argument."
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
}

// Checks if all preconditions are met, before actually executing
// our business function.
func validateUpgradeClusterPreconditions(args Arguments, cmdLineArgs []string) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	// authentication
	if config.Config.Token == "" && args.AuthToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	// cluster ID is present
	if args.ClusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	}

	return nil
}

// upgradeClusterExecutionOutput executes our business function and displays the result,
// both in case of success or error
func upgradeClusterExecutionOutput(cmd *cobra.Command, cmdLineArgs []string) {
	result, err := upgradeCluster(arguments)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case err.Error() == "":
			return
		case errors.IsCouldNotCreateClientError(err):
			headline = "Failed to create API client."
			subtext = "Details: " + err.Error()
		case errors.IsNoUpgradeAvailableError(err):
			headline = "There is no newer release available."
			subtext = "Please check the available releases using 'gsctl list releases'."
		case errors.IsClusterNotFoundError(err):
			headline = "The cluster does not exist."
			subtext = fmt.Sprintf("We couldn't find a cluster '%s' via API endpoint %s.", arguments.ClusterNameOrID, arguments.APIEndpoint)
		case errors.IsCommandAbortedError(err):
			headline = "Not upgrading."
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

	fmt.Println(color.GreenString("Starting to upgrade cluster '%s' to release version %s",
		result.clusterID,
		result.versionAfter))
}

// upgradeCluster performs our actual function. It usually creates an API client,
// configures it, configures an API request and performs it.
func upgradeCluster(args Arguments) (*upgradeClusterResult, error) {
	result := &upgradeClusterResult{}

	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result.clusterID, err = clustercache.GetID(args.APIEndpoint, args.ClusterNameOrID, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = upgradeClusterActivityName

	// Fetch cluster details, detect API version to use.
	var detailsV4 *models.V4ClusterDetailsResponse
	var detailsV5 *models.V5ClusterDetailsResponse
	{
		if args.Verbose {
			fmt.Println(color.WhiteString("Attempt to fetch v5 cluster details."))
		}

		responseV5, v5err := clientWrapper.GetClusterV5(result.clusterID, auxParams)
		if errors.IsClusterNotFoundError(v5err) || clienterror.IsBadRequestError(v5err) {
			if args.Verbose {
				fmt.Println(color.WhiteString("Not found via v5 endpoint. Attempt to fetch v4 cluster details."))
			}

			responseV4, v4err := clientWrapper.GetClusterV4(result.clusterID, auxParams)
			if v4err != nil {
				return nil, microerror.Mask(v4err)
			}

			detailsV4 = responseV4.Payload
			result.versionBefore = detailsV4.ReleaseVersion
		} else if v5err != nil {
			return nil, microerror.Mask(v5err)
		} else {
			detailsV5 = responseV5.Payload
			result.versionBefore = detailsV5.ReleaseVersion
		}
	}

	releasesResponse, err := clientWrapper.GetReleases(auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var releaseVersions []string
	for _, r := range releasesResponse.Payload {
		// filter out non-active releases
		if !r.Active || !isVersionProductionReady(*r.Version) {
			continue
		}

		releaseVersions = append(releaseVersions, *r.Version)
	}

	// define the target version to upgrade to
	if args.Verbose {
		fmt.Println(color.WhiteString("Obtaining information on the successor release."))
	}

	var targetVersion string
	{
		if args.Release == "" {
			targetVersion = successorReleaseVersion(result.versionBefore, releaseVersions)
			if targetVersion == "" {
				return nil, microerror.Mask(errors.NoUpgradeAvailableError)
			}
		} else {
			targetVersion = args.Release
		}
	}

	result.versionAfter = targetVersion

	var targetRelease models.V4ReleaseListItem
	for _, rel := range releasesResponse.Payload {
		if *rel.Version == targetVersion {
			targetRelease = *rel
		}
	}

	if targetRelease.Version == nil {
		// Release was not found.
		return nil, microerror.Maskf(errors.InvalidReleaseError, fmt.Sprintf("Can't upgrade to non existing release %s", targetVersion))
	}

	// Show some details independent of confirmation
	if !targetRelease.Active {
		fmt.Printf("Cluster '%s' will be upgraded from version %s to %s, which is not an active release.\n",
			color.CyanString(args.ClusterNameOrID),
			color.CyanString(result.versionBefore),
			color.CyanString(targetVersion))
		fmt.Printf("This might fail depending on your permissions.\n")
	} else {
		fmt.Printf("Cluster '%s' will be upgraded from version %s to %s.\n",
			color.CyanString(args.ClusterNameOrID),
			color.CyanString(result.versionBefore),
			color.CyanString(targetVersion))
	}

	// Details output
	fmt.Println("")
	fmt.Println("Changelog:")
	fmt.Println("")

	for _, change := range targetRelease.Changelog {
		fmt.Printf("    - %s: %s\n", change.Component, change.Description)
	}

	fmt.Println("")
	fmt.Println("NOTE: Upgrading may impact your running workloads and will make the cluster's")
	fmt.Println("Kubernetes API unavailable temporarily. Before upgrading, please acknowledge the")
	fmt.Println("details described in")
	fmt.Println("")
	fmt.Printf("    %s\n", upgradeDocsURL)
	fmt.Println("")

	// Confirmation
	if !args.Force {
		confirmed := confirm.Ask("Do you want to start the upgrade now?")
		if !confirmed {
			return nil, microerror.Mask(errors.CommandAbortedError)
		}
	}

	if detailsV5 != nil {
		if args.Verbose {
			fmt.Println(color.WhiteString("Submitting cluster modification request to v5 endpoint."))
		}

		reqBody := &models.V5ModifyClusterRequest{
			ReleaseVersion: targetVersion,
		}

		_, err = clientWrapper.ModifyClusterV5(result.clusterID, reqBody, auxParams)
		if err != nil {
			return nil, microerror.Maskf(errors.CouldNotUpgradeClusterError, err.Error())
		}
	} else {
		if args.Verbose {
			fmt.Println(color.WhiteString("Submitting cluster modification request to v4 endpoint."))
		}

		reqBody := &models.V4ModifyClusterRequest{
			ReleaseVersion: targetVersion,
		}

		// perform API call
		_, err = clientWrapper.ModifyClusterV4(result.clusterID, reqBody, auxParams)
		if err != nil {
			return nil, microerror.Maskf(errors.CouldNotUpgradeClusterError, err.Error())
		}
	}

	return result, nil
}

func isVersionProductionReady(version string) bool {
	semverVersion, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	return len(semverVersion.Prerelease()) < 1 && len(semverVersion.Metadata()) < 1
}

// successorReleaseVersion returns the lowest version number from a slice
// that is still higher than the comparison version.
// If the current version is empty or no successor is found, returns an empty string.
func successorReleaseVersion(version string, versions []string) string {
	if version == "" {
		return ""
	}

	// sort versions by semver number
	sort.Slice(versions, func(i, j int) bool {
		return util.VersionSortComp(versions[i], versions[j])
	})

	// return first item that is greater than version
	for _, v := range versions {
		comp, _ := util.CompareVersions(v, version)
		if comp == 1 {
			return v
		}
	}

	return ""
}
