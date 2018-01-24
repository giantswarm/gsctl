package commands

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/microerror"
)

var (
	// UpgradeClusterCommand performs the "upgrade cluster" function
	UpgradeClusterCommand = &cobra.Command{
		Use:   "cluster",
		Short: "Upgrades a cluster to a newer release version",
		Long: `Upgrades a cluster to a newer release version.

Upgrades mean the stepwise replacement of the workers, the master and other
building blocks of a cluster with newer versions.

A cluster will always be upgraded to the subsequent release. To find out what
release version is used currently, use

    gsctl show cluster -c <cluster-id>

To find out what is the subsequent version, list all available versions using

    gsctl list releases

TODO:
- Explain more
- Link to in-depth docs article`,

		// We use PreRun for general input validation, authentication etc.
		// If something is bad/missing, that function has to exit with a
		// non-zero exit code.
		PreRun: upgradeClusterPreRunOutput,

		// Run is the function that actually executes what we want to do.
		Run: upgradeClusterRunOutput,
	}
)

const (
	// upgradeClusterActivityName assigns API requests to named activities
	upgradeClusterActivityName = "upgrade-cluster"
)

// argument struct to pass to our business function and
// to the validation function
type upgradeClusterArguments struct {
	apiEndpoint string
	authToken   string
	clusterID   string
	force       bool
	verbose     bool
}

// function to create arguments based on command line flags and config
func defaultUpgradeClusterArguments(cmdLineArgs []string) upgradeClusterArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	clusterID := ""
	if len(cmdLineArgs) > 0 {
		clusterID = cmdLineArgs[0]
	}

	return upgradeClusterArguments{
		apiEndpoint: endpoint,
		authToken:   token,
		clusterID:   clusterID,
		force:       false,
		verbose:     cmdVerbose,
	}
}

type upgradeClusterResult struct {
	clusterID     string
	versionBefore string
	versionAfter  string
}

// Here we populate our cobra command
func init() {
	UpgradeClusterCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, no interactive confirmation will be required (risky!).")

	UpgradeCommand.AddCommand(UpgradeClusterCommand)
}

// Prints results of our pre-validation
func upgradeClusterPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultUpgradeClusterArguments(cmdLineArgs)
	err := verifyUpgradeClusterPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	headline := ""
	subtext := ""

	switch {
	case err.Error() == "":
		return
	case IsNotLoggedInError(err):
		headline = "You are not logged in."
		subtext = fmt.Sprintf("Use '%s login' to login or '--auth-token' to pass a valid auth token.", config.ProgramName)
	case IsClusterIDMissingError(err):
		headline = "No cluster ID specified."
		subtext = "Please specify which cluster to upgrade by using the cluster ID as an argument."
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

// Checks if all preconditions are met, before actually executing
// our business function
func verifyUpgradeClusterPreconditions(args upgradeClusterArguments, cmdLineArgs []string) error {
	// authentication
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}

	// cluster ID is present
	if args.clusterID == "" {
		return microerror.Mask(clusterIDMissingError)
	}

	return nil
}

// upgradeClusterRunOutput executes our business function and displays the result,
// both in case of success or error
func upgradeClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultUpgradeClusterArguments(cmdLineArgs)
	result, err := upgradeCluster(args)

	if err != nil {
		var headline = ""
		var subtext = ""

		switch {
		case err.Error() == "":
			return
		case IsCouldNotCreateClientError(err):
			headline = "Failed to create API client."
			subtext = "Details: " + err.Error()
		case IsNoUpgradeAvailableError(err):
			headline = "There is no newer release available."
			subtext = "Please check the available releases using 'gsctl list releases'."
		case IsClusterNotFoundError(err):
			headline = "The cluster does not exist."
			subtext = fmt.Sprintf("We couldn't find a cluster with the ID '%s' via API endpoint %s.", args.clusterID, args.apiEndpoint)
		case IsCommandAbortedError(err):
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

	fmt.Printf("Starting to upgrade cluster %s to release version %s", result.clusterID, result.versionAfter)
}

// upgradeCluster performs our actual function. It usually creates an API client,
// configures it, configures an API request and performs it.
func upgradeCluster(args upgradeClusterArguments) (upgradeClusterResult, error) {
	result := upgradeClusterResult{}

	// fetch current cluster details
	details, detailsErr := getClusterDetails(args.clusterID, args.authToken, args.apiEndpoint)
	if detailsErr != nil {
		return result, microerror.Mask(detailsErr)
	}

	listReleasesArgs := listReleasesArguments{
		apiEndpoint: args.apiEndpoint,
		token:       args.authToken,
	}
	releasesResult, releasesErr := listReleases(listReleasesArgs)
	if releasesErr != nil {
		return result, microerror.Mask(releasesErr)
	}
	releaseVersions := []string{}
	for _, r := range releasesResult.releases {
		releaseVersions = append(releaseVersions, r.Version)
	}

	newVersion := successorReleaseVersion(details.ReleaseVersion, releaseVersions)
	if newVersion == "" {
		return result, microerror.Mask(noUpgradeAvailableError)
	}

	// confirmation
	if !args.force {
		// Show information before confirmation
		fmt.Printf("Cluster '%s' will be upgraded from version %s to %s.\n", args.clusterID, details.ReleaseVersion, newVersion)

		fmt.Println("")
		fmt.Println("Changelog:")
		fmt.Println("")

		for _, release := range releasesResult.releases {
			if release.Version == newVersion {
				for _, change := range release.Changelog {
					fmt.Printf("    - %s: %s\n", change.Component, change.Description)
				}
			}
		}

		fmt.Println("")
		fmt.Println("NOTE: Upgrading may impact your running workloads and will make the cluster's")
		fmt.Println("Kubernetes API unavailable temporarily. Before upgrading, please acknowledge the")
		fmt.Println("details described in")
		fmt.Println("")
		fmt.Println("    https://docs.giantswarm.io/reference/cluster-upgrades/")
		fmt.Println("")

		confirmed := askForConfirmation("Do you want to start the upgrade now?")
		if !confirmed {
			return result, microerror.Mask(commandAbortedError)
		}
	}

	result.clusterID = args.clusterID
	result.versionBefore = details.ReleaseVersion

	// create API client
	authHeader := "giantswarm " + config.Config.Token
	if args.authToken != "" {
		// command line flag overwrites
		authHeader = "giantswarm " + args.authToken
	}
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(couldNotCreateClientError)
	}

	// request body
	reqBody := gsclientgen.V4ModifyClusterRequest{
		ReleaseVersion: newVersion,
	}

	// perform API call
	_, rawResponse, err := apiClient.ModifyCluster(authHeader, args.clusterID, reqBody, requestIDHeader, upgradeClusterActivityName, cmdLine)
	if err != nil {
		return result, microerror.Mask(err)
	}

	if rawResponse.Response.StatusCode != http.StatusOK {
		// error response with code/message body
		genericResponse, err := client.ParseGenericResponse(rawResponse.Payload)
		if err == nil {
			if args.verbose {
				fmt.Printf("\nError details:\n - Code: %s\n - Message: %s\n\n",
					genericResponse.Code, genericResponse.Message)
			}
			return result, microerror.Mask(couldNotUpgradeClusterError)
		}

		// other response body format
		if args.verbose {
			fmt.Printf("\nError details:\n - HTTP status code: %d\n - Response body: %s\n\n",
				rawResponse.Response.StatusCode,
				string(rawResponse.Payload))
		}
		return result, microerror.Mask(couldNotScaleClusterError)
	}

	result.versionAfter = newVersion

	return result, nil
}

// successorReleaseVersion returns the lowest version number from a slice
// that is still higher than the comparison version.
func successorReleaseVersion(version string, versions []string) string {
	// sort versions by semver number
	sort.Slice(versions, func(i, j int) bool {
		vi := semver.New(versions[i])
		vj := semver.New(versions[j])
		return vi.LessThan(*vj)
	})

	// return first item that is greater than version
	comp := semver.New(version)
	for _, v := range versions {
		vv := semver.New(v)
		if comp.LessThan(*vv) {
			return v
		}
	}

	return ""
}
