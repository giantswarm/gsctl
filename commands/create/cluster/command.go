/*
Package cluster defines the 'create cluster' command.

The command deals with a few delicacies/spiecialties:

- Cluster spec details, e. g. which instance type to use for workers, can
  be specified by the user, but don't have to. The backend will
  fill in missing details using defaulting.

- Cluster spec details can be provided either using command line flags
  or by passing a YAML definition. When passing a YAML definition, some
  attributes from that definition can even be overridden using flags.

- On AWS, starting from a certain release version, clusters will have
  node pools and will be created using the v5 API endpoint. On other providers
  as well as on AWS for older releases, the v4 API endpoint has to be used.

*/
package cluster

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/juju/errgo"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/capabilities"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/limits"
)

// Arguments contains all possible input parameter needed
// (and optionally available) for creating a cluster.
type Arguments struct {
	APIEndpoint       string
	AuthToken         string
	ClusterName       string
	FileSystem        afero.Fs
	InputYAMLFile     string
	Owner             string
	ReleaseVersion    string
	Scheme            string
	UserProvidedToken string
	Verbose           bool
}

// collectArguments gets arguments from flags and returns an Arguments object.
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		APIEndpoint:       endpoint,
		AuthToken:         token,
		ClusterName:       flags.ClusterName,
		FileSystem:        config.FileSystem,
		InputYAMLFile:     flags.InputYAMLFile,
		Owner:             flags.Owner,
		ReleaseVersion:    flags.Release,
		Scheme:            scheme,
		UserProvidedToken: flags.Token,
		Verbose:           flags.Verbose,
	}
}

// creationResult is the struct to gather all our API call results.
type creationResult struct {
	// cluster ID
	ID string
	// location to fetch details on new cluster from
	Location string
	// cluster definition assembled, v4 compatible
	DefinitionV4 *types.ClusterDefinitionV4
	DefinitionV5 *types.ClusterDefinitionV5

	// HasErrors should be true if we saw some non-critical errors.
	// This is only relevant in v5 and should only be used if a node
	// pool could not be created successfully.
	HasErrors bool
}

const (
	createClusterActivityName = "create-cluster"
)

var (
	// Command performs the "create cluster" function
	Command = &cobra.Command{
		Use:   "cluster",
		Short: "Create cluster",
		Long: `Creates a new Kubernetes cluster.

You can specify a few cluster attributes like name, owner and release version
using command line flags. Additional attributes regarding the worker node
specification can be added using a YAML definition file.

For details about the cluster definition YAML format, see

  https://docs.giantswarm.io/reference/cluster-definition/

Note that you can also command line flags override settings from the YAML
definition.

Defaults
--------

All parameters you don't set explicitly will be set using defaults. You can
get some information on these defaults using the 'gsctl info' command, as they
might be specific for the installation you are working with. Here are some
general defaults:

- Release: the latest release is used.
- Workers
  - On AWS and when using the latest release, the cluster will be created
    without any node pools. You may define node pools in your cluster
    definition YAML or add node pools one by one using 'gsctl create nodepool'.
  - On AWS with releases prior to node pools, and with Azure and KVM, the
    cluster will have three worker nodes by default, using pretty much the
	minimal spec for a working cluster.
  - Autoscaling will be inactive initially, as the minimum and maximum of the
    scaling range  will be set to 3.
  - All worker nodes will be in the same availability zone.
  - The cluster will have a generic name.

Examples:

  gsctl create cluster --owner acme

  gsctl create cluster --owner myorg --name "Production Cluster"

  gsctl create cluster --file ./cluster.yaml

  gsctl create cluster --file ./staging-cluster.yaml \
    --owner acme --name Staging

  cat my-cluster.yaml | gsctl create cluster -f -

  `,
		PreRun: printValidation,
		Run:    printResult,
	}

	// the client wrapper we will use in this command.
	clientWrapper *client.Wrapper

	// nodePoolsEnabled stores whether we can assume the v5 API (node pools) for this command execution.
	nodePoolsEnabled = false
)

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().StringVarP(&flags.InputYAMLFile, "file", "f", "", "Path to a cluster definition YAML file. Use '-' to read from STDIN.")
	Command.Flags().StringVarP(&flags.ClusterName, "name", "n", "", "Cluster name")
	Command.Flags().StringVarP(&flags.Owner, "owner", "o", "", "Organization to own the cluster")
	Command.Flags().StringVarP(&flags.Release, "release", "r", "", "Release version to use, e. g. '1.2.3'. Defaults to the latest. See 'gsctl list releases --help' for details.")
}

// printValidation runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func printValidation(cmd *cobra.Command, positionalArgs []string) {
	args := collectArguments()

	headline := ""
	subtext := ""

	err := verifyPreconditions(args)
	if err != nil {
		errors.HandleCommonErrors(err)

		switch {
		case errors.IsConflictingFlagsError(err):
			headline = "Conflicting flags used"
			subtext = "When specifying a definition via a YAML file, certain flags must not be used."
		case errors.IsConflictingWorkerFlagsUsed(err):
			headline = "Conflicting flags used"
			subtext = "When specifying --num-workers, neither --workers-max nor --workers-min must be used."
		case errors.IsWorkersMinMaxInvalid(err):
			headline = "Number of worker nodes invalid"
			subtext = "Node count flag --workers-min must not be higher than --workers-max."
		case errors.IsNumWorkerNodesMissingError(err):
			headline = "Number of worker nodes required"
			subtext = "When specifying worker node details, you must also specify the number of worker nodes."
		case errors.IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("You'll need at least %v worker nodes for a useful cluster.", limits.MinimumNumWorkers)
		case errors.IsNotEnoughCPUCoresPerWorkerError(err):
			headline = "Not enough CPUs per worker specified"
			subtext = fmt.Sprintf("You'll need at least %v CPU cores per worker node.", limits.MinimumWorkerNumCPUs)
		case errors.IsNotEnoughMemoryPerWorkerError(err):
			headline = "Not enough Memory per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", limits.MinimumWorkerMemorySizeGB)
		case errors.IsNotEnoughStoragePerWorkerError(err):
			headline = "Not enough Storage per worker specified"
			subtext = fmt.Sprintf("You'll need at least %.1f GB per worker node.", limits.MinimumWorkerStorageSizeGB)
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

// printResult calls addCluster() and creates user-friendly output of the result
func printResult(cmd *cobra.Command, positionalArgs []string) {
	// use arguments as passed from command line via cobra
	args := collectArguments()

	result, err := addCluster(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		var headline string
		var subtext string
		richError, richErrorOK := err.(*errgo.Err)

		switch {
		case errors.IsClusterOwnerMissingError(err):
			headline = "No owner organization set"
			subtext = "Please specify an owner organization for the cluster via the --owner flag."
			if args.InputYAMLFile != "" {
				subtext = "Please specify an owner organization for the cluster in your definition file or set one via the --owner flag."
			}
		case errors.IsNotEnoughWorkerNodesError(err):
			headline = "Not enough worker nodes specified"
			subtext = fmt.Sprintf("If you specify workers in your definition file, you'll have to specify at least %d worker nodes for a useful cluster.", limits.MinimumNumWorkers)
		case errors.IsYAMLNotParseable(err):
			headline = "Could not parse YAML"
			if args.InputYAMLFile == "-" {
				subtext = "The YAML data given via STDIN could not be parsed into a cluster definition."
			} else {
				subtext = fmt.Sprintf("The YAML data read from file '%s' could not be parsed into a cluster definition.", args.InputYAMLFile)
			}
		case errors.IsYAMLFileNotReadable(err):
			headline = "Could not read YAML file"
			subtext = fmt.Sprintf("The file '%s' could not be read. Please make sure that it is readable and contains valid YAML.", args.InputYAMLFile)
		case errors.IsCouldNotCreateJSONRequestBodyError(err):
			headline = "Could not create the JSON body for cluster creation API request"
			subtext = "There seems to be a problem in parsing the cluster definition. Please contact Giant Swarm via Slack or via support@giantswarm.io with details on how you executes this command."
		case errors.IsNotAuthorizedError(err):
			headline = "Not authorized"
			subtext = "No cluster has been created, as you are are not authenticated or not authorized to perform this action."
			subtext += " Please check your credentials or, to make sure, use 'gsctl login' to log in again."
		case errors.IsOrganizationNotFoundError(err):
			headline = "Organization not found"
			subtext = "The organization set to own the cluster does not exist."
		case errors.IsCouldNotCreateClusterError(err):
			headline = "The cluster could not be created."
			subtext = "You might try again in a few moments. If that doesn't work, please contact the Giant Swarm support team."
			subtext += " Sorry for the inconvenience!"

			// more details for backend side / connection errors
			subtext += "\n\nDetails:\n"
			if richErrorOK {
				subtext += richError.Message()
			} else {
				subtext += err.Error()
			}

		default:
			headline = err.Error()
		}

		// output error information
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	// success output
	if result.DefinitionV4.Name != "" {
		fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' is launching.", result.DefinitionV4.Name, result.ID, result.DefinitionV4.Owner))
	} else {
		fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' is launching.", result.ID, result.DefinitionV4.Owner))
	}
	fmt.Println("Add key pair and settings to kubectl using")
	fmt.Println("")
	fmt.Printf("    %s", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s \n", result.ID)))
	fmt.Println("")
	fmt.Println("Take into consideration all clusters have enabled RBAC and may you want to provide a correct organization for the certificates (like operators, testers, developer, ...)")
	fmt.Println("")
	fmt.Printf("    %s \n", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s --certificate-organizations system:masters", result.ID)))
	fmt.Println("")
	fmt.Println("To know more about how to create the kubeconfig run")
	fmt.Println("")
	fmt.Printf("    %s \n\n", color.YellowString("gsctl create kubeconfig --help"))
}

// verifyPreconditions checks preconditions and returns an error in case.
func verifyPreconditions(args Arguments) error {
	// logged in?
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}

	return nil
}

// getLatestActiveReleaseVersion returns the latest active release.
func getLatestActiveReleaseVersion(clientWrapper *client.Wrapper, auxParams *client.AuxiliaryParams) (string, error) {
	response, err := clientWrapper.GetReleases(auxParams)
	if err != nil {
		return "", microerror.Mask(err)
	}

	activeReleases := []*models.V4ReleaseListItem{}
	for _, r := range response.Payload {
		if r.Active {
			activeReleases = append(activeReleases, r)
		}
	}

	// sort releases by version (descending)
	sort.Slice(activeReleases[:], func(i, j int) bool {
		vi, err := semver.NewVersion(*activeReleases[i].Version)
		if err != nil {
			return false
		}
		vj, err := semver.NewVersion(*activeReleases[j].Version)
		if err != nil {
			return true
		}

		return vi.GreaterThan(vj)
	})

	return *activeReleases[0].Version, nil
}

// addCluster collects information to decide whether to create a cluster
// via the v4 or v5 API endpoint, then returns results.
func addCluster(args Arguments) (*creationResult, error) {
	var err error

	clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = createClusterActivityName

	// Ensure provider information is there.
	if config.Config.Provider == "" {
		if flags.Verbose {
			fmt.Println(color.WhiteString("Fetching installation information"))
		}

		info, err := clientWrapper.GetInfo(auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		config.Config.SetProvider(info.Payload.General.Provider)
	}

	// Process YAML definition (if given), so we can take a 'release_version' key into consideration.
	var definitionInterface interface{}
	if args.InputYAMLFile == "-" {
		definitionInterface, err = readDefinitionFromSTDIN()

		if err != nil {
			return nil, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
		}
	} else if args.InputYAMLFile != "" {
		// definition from file (and optionally flags)
		definitionInterface, err = readDefinitionFromFile(args.FileSystem, args.InputYAMLFile)
		if err != nil {
			return nil, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
		}
	}

	var wantedRelease string
	if config.Config.Provider == "aws" {
		if args.ReleaseVersion != "" {
			wantedRelease = args.ReleaseVersion
		} else {
			// look at release version from YAML definition
			if defV5, okV5 := definitionInterface.(types.ClusterDefinitionV5); okV5 {
				if defV5.ReleaseVersion != "" {
					wantedRelease = defV5.ReleaseVersion
				}
			}
			if defV4, okV4 := definitionInterface.(types.ClusterDefinitionV4); okV4 {
				if defV4.ReleaseVersion != "" {
					wantedRelease = defV4.ReleaseVersion
				}
			}
		}

		// As no other is set, use latest active release.
		if wantedRelease == "" {
			latest, err := getLatestActiveReleaseVersion(clientWrapper, auxParams)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			wantedRelease = latest
		}

		// Fetch node pools capabilities info.
		capabilityService, err := capabilities.New(config.Config.Provider, clientWrapper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		nodePoolsEnabled, err = capabilityService.HasCapability(wantedRelease, capabilities.NodePools)
	}

	result := &creationResult{}

	if definitionInterface != nil {
		if def, ok := definitionInterface.(*types.ClusterDefinitionV5); ok {
			result.DefinitionV5 = def
			nodePoolsEnabled = true
		} else if def, ok := definitionInterface.(*types.ClusterDefinitionV4); ok {
			result.DefinitionV4 = def
		} else {
			return nil, microerror.Mask(errors.YAMLNotParseableError)
		}
	}

	if nodePoolsEnabled {
		if args.Verbose {
			fmt.Println(color.WhiteString("Using the v5 API to create a cluster with node pool support"))
		}

		if result.DefinitionV5 == nil {
			result.DefinitionV5 = &types.ClusterDefinitionV5{}
		}

		updateDefinitionFromFlagsV5(result.DefinitionV5, args.ClusterName, args.ReleaseVersion, args.Owner)

		id, hasErrors, err := addClusterV5(result.DefinitionV5, args, clientWrapper, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result.ID = id
		result.HasErrors = hasErrors

	} else {
		if args.Verbose {
			fmt.Println(color.WhiteString("Using the v4 API to create a cluster"))
		}

		if result.DefinitionV4 == nil {
			result.DefinitionV4 = &types.ClusterDefinitionV4{}
		}

		updateDefinitionFromFlagsV4(result.DefinitionV4, args.ClusterName, args.ReleaseVersion, args.Owner)

		id, location, err := addClusterV4(result.DefinitionV4, args, clientWrapper, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result.ID = id
		result.Location = location
	}

	return result, nil
}
