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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/juju/errgo"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/capabilities"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
)

// Arguments contains all possible input parameter needed
// (and optionally available) for creating a cluster.
type Arguments struct {
	APIEndpoint           string
	AuthToken             string
	CreateDefaultNodePool bool
	ClusterName           string
	FileSystem            afero.Fs
	InputYAMLFile         string
	Owner                 string
	ReleaseVersion        string
	Scheme                string
	MasterHA              *bool
	UserProvidedToken     string
	Verbose               bool
	OutputFormat          string
}

// collectArguments gets arguments from flags and returns an Arguments object.
func collectArguments(cmd *cobra.Command) Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	var haMasters *bool
	if cmd.Flag("master-ha").Changed {
		haMasters = &flags.MasterHA
	}

	return Arguments{
		APIEndpoint:           endpoint,
		AuthToken:             token,
		ClusterName:           flags.ClusterName,
		CreateDefaultNodePool: flags.CreateDefaultNodePool,
		FileSystem:            config.FileSystem,
		InputYAMLFile:         flags.InputYAMLFile,
		MasterHA:              haMasters,
		Owner:                 flags.Owner,
		ReleaseVersion:        flags.Release,
		Scheme:                scheme,
		UserProvidedToken:     flags.Token,
		Verbose:               flags.Verbose,
		OutputFormat:          flags.OutputFormat,
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

// JSONOutput contains the fields included in JSON output of the create cluster command when called with json output flag
type JSONOutput struct {
	// cluster ID
	ID string `json:"id,omitempty"`
	// result. should be 'created'
	Result string `json:"result"`
	// Error which occured
	Error error `json:"error,omitempty"`
}

const (
	createClusterActivityName = "create-cluster"

	standardInputSpecialPath = "-"
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
  - On AWS and when using the latest release, and when not specifying node pool
    details via a cluster definition, the cluster will be created with a
    default node pool. You may define node pools in your cluster definition
    YAML or add node pools one by one using 'gsctl create nodepool'.
  - On AWS with releases prior to node pools, and with Azure and KVM, the
    cluster will have three worker nodes by default, using pretty much the
    minimal spec for a working cluster.
  - Autoscaling will be inactive initially, as the minimum and maximum of the
    scaling range  will be set to 3.
  - All worker nodes will be in the same availability zone.
  - The cluster will have a generic name.

Examples:

  gsctl create cluster --owner acme

  gsctl create cluster --owner acme --name "Production Cluster"

  gsctl create cluster --file ./cluster.yaml

  gsctl create cluster --file ./staging-cluster.yaml \
    --owner acme --name Staging

  cat my-cluster.yaml | gsctl create cluster -f -

With Bash and other compatible shells, the syntax shown below can be used to
create a YAML defininition and pass it to the command in one go, without the
need for a file:

gsctl create cluster -f - <<EOF
owner: acme
name: Test cluster using two AZs
release_version: 8.2.0
availability_zones: 2
EOF

For v5 clusters (those with node pool support) gsctl automatically creates a
node pool using default settings, if you don't specify your own node pools.
You can suppress the creation of the default node pool by setting the
flag --create-default-nodepool to false. Example:

  gsctl create cluster \
    --owner acme \
    --name "Empty cluster" \
    --create-default-nodepool=false
`,
		PreRun: printValidation,
		Run:    printResult,
	}

	// the client wrapper we will use in this command.
	clientWrapper *client.Wrapper

	arguments Arguments
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
	Command.Flags().BoolVar(&flags.MasterHA, "master-ha", true, "This means the cluster will have three master nodes. Requires High-Availability Master support.")
	Command.Flags().BoolVarP(&flags.CreateDefaultNodePool, "create-default-nodepool", "", true, "Whether a default node pool should be created if none is specified in the definition. Requires node pool support.")
	Command.Flags().StringVarP(&flags.OutputFormat, "output", "", "", fmt.Sprintf("Output format. Specifying '%s' will change output to be JSON formatted.", formatting.OutputFormatJSON))
}

// printValidation runs our pre-checks.
// If errors occur, error info is printed to STDOUT/STDERR
// and the program will exit with non-zero exit codes.
func printValidation(cmd *cobra.Command, positionalArgs []string) {
	arguments = collectArguments(cmd)

	headline := ""
	subtext := ""

	err := verifyPreconditions(arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		switch {
		case errors.IsConflictingFlagsError(err):
			headline = "Conflicting flags used"
			subtext = "When specifying a definition via a YAML file, certain flags must not be used."
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
	result, err := addCluster(arguments)

	if arguments.OutputFormat == formatting.OutputFormatJSON {
		fmt.Println(getJSONOutput(result, err))
		return
	}

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline string
		var subtext string
		richError, richErrorOK := err.(*errgo.Err)

		switch {
		case IsHAMastersNotSupported(err):
			var haMastersRequiredVersion string
			{
				for _, requiredRelease := range capabilities.HAMasters.RequiredReleasePerProvider {
					if requiredRelease.Provider == config.Config.Provider {
						haMastersRequiredVersion = requiredRelease.ReleaseVersion.String()

						break
					}
				}
			}

			headline = "Feature not supported"

			if haMastersRequiredVersion == "" {
				subtext = fmt.Sprintf("Master node high availability is not supported by your provider. (%s)", strings.ToUpper(config.Config.Provider))
			} else {
				subtext = fmt.Sprintf("Master node high availability is only supported by releases %s and higher.", haMastersRequiredVersion)
			}
		case IsMustProvideSingleMasterType(err):
			headline = "Conflicting master node configuration"
			subtext = "The release version you're trying to use supports master node high availability.\nPlease remove the 'master' attribute from your cluster definition and use the 'master_nodes' attribute instead."
		case errors.IsClusterOwnerMissingError(err):
			headline = "No owner organization set"
			subtext = "Please specify an owner organization for the cluster via the --owner flag."
			if arguments.InputYAMLFile != "" {
				subtext = "Please specify an owner organization for the cluster in your definition file or set one via the --owner flag."
			}
		case errors.IsYAMLNotParseable(err):
			headline = "Could not parse YAML"
			if arguments.InputYAMLFile == standardInputSpecialPath {
				subtext = "The YAML data given via STDIN could not be parsed into a cluster definition."
			} else {
				subtext = fmt.Sprintf("The YAML data read from file '%s' could not be parsed into a cluster definition.", arguments.InputYAMLFile)
			}
		case errors.IsYAMLFileNotReadable(err):
			if arguments.InputYAMLFile == standardInputSpecialPath {
				headline = "Could not read YAML from STDIN"
				subtext = "The YAML definition given via standard input could not be parsed.\n"
				subtext += fmt.Sprintf("Details: %s", err.Error())
			} else {
				headline = "Could not read YAML file"
				subtext = fmt.Sprintf("The file '%s' could not be read. Please make sure that it is readable and contains valid YAML.\n", arguments.InputYAMLFile)
				subtext += fmt.Sprintf("Details: %s", err.Error())
			}
		case errors.IsIncompatibleSettings(err):
			headline = "Incompatible settings"
			subtext = "The provided cluster details/definition are not compatible with the capabilities of the installation and/or release.\n"
			subtext += fmt.Sprintf("Error details: %s", err.Error())
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
	if result.DefinitionV4 != nil {
		if result.DefinitionV4.Name != "" {
			fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' is launching.", result.DefinitionV4.Name, result.ID, result.DefinitionV4.Owner))
		} else {
			fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' is launching.", result.ID, result.DefinitionV4.Owner))
		}
	} else if result.DefinitionV5 != nil {
		if result.DefinitionV5.Name != "" {
			fmt.Println(color.GreenString("New cluster '%s' (ID '%s') for organization '%s' has been created.", result.DefinitionV5.Name, result.ID, result.DefinitionV5.Owner))
		} else {
			fmt.Println(color.GreenString("New cluster with ID '%s' for organization '%s' has been created.", result.ID, result.DefinitionV5.Owner))
		}

		if result.HasErrors {
			fmt.Println("Note: Some error(s) occurred during node pool creation. Please check the error details above.")
			fmt.Printf("To verify that nodes are coming up, please use the following commands:\n\n")
			fmt.Printf("    %s \n", color.YellowString(fmt.Sprintf("gsctl list nodepools %s", result.ID)))
			fmt.Printf("    %s \n", color.YellowString(fmt.Sprintf("gsctl show nodepool %s/<nodepool-id>", result.ID)))
		}
	}

	fmt.Println("\nAdd a key pair and settings for kubectl using")
	fmt.Println("")
	fmt.Printf("    %s", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s \n", result.ID)))
	fmt.Println("")
	fmt.Println("Take into consideration all clusters have enabled RBAC and may you want to provide a correct organization for the certificates (like operators, testers, developer, ...).")
	fmt.Println("")
	fmt.Printf("    %s \n", color.YellowString(fmt.Sprintf("gsctl create kubeconfig --cluster=%s --certificate-organizations system:masters", result.ID)))
	fmt.Println("")
	fmt.Println("To know more about how to create the kubeconfig run")
	fmt.Println("")
	fmt.Printf("    %s \n\n", color.YellowString("gsctl create kubeconfig --help"))
}

func getJSONOutput(result *creationResult, creationErr error) string {
	var outputBytes []byte
	var err error
	var jsonResult JSONOutput

	// handle errors
	if creationErr != nil {
		jsonResult = JSONOutput{Result: "error", Error: creationErr}
	} else {
		jsonResult = JSONOutput{ID: result.ID, Result: "created"}
		if result.HasErrors {
			jsonResult.Result = "created-with-errors"
		}
	}

	outputBytes, err = json.MarshalIndent(jsonResult, formatting.OutputJSONPrefix, formatting.OutputJSONIndent)
	if err != nil {
		os.Exit(1)
	}

	return string(outputBytes)
}

// verifyPreconditions checks preconditions and returns an error in case.
func verifyPreconditions(args Arguments) error {
	if args.APIEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	// logged in?
	if args.AuthToken == "" && args.UserProvidedToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.OutputFormat != "" && args.OutputFormat != formatting.OutputFormatJSON {
		return microerror.Maskf(errors.OutputFormatInvalidError, fmt.Sprintf("Output format '%s' is unknown. Valid options: '%s'", args.OutputFormat, formatting.OutputFormatJSON))
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
// via the v4 or v5 API endpoint, then calls the according functions
// and returns results.
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
		if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
			fmt.Println(color.WhiteString("Fetching installation information"))
		}

		info, err := clientWrapper.GetInfo(auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = config.Config.SetProvider(info.Payload.General.Provider)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Process YAML definition (if given), so we can take a 'release_version' key into consideration.
	var definitionInterface interface{}
	if args.InputYAMLFile == standardInputSpecialPath {
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

	// The release version we are selecting, based on command line flags, YAML definition,
	// or as the latest release available.
	var wantedRelease string

	// nodePoolsEnabled stores whether we can assume the v5 API (node pools) for this command execution.
	var nodePoolsEnabled bool
	var haMastersEnabled bool

	var usesV4Definition, usesV5Definition bool
	var defV4 types.ClusterDefinitionV4
	var defV5 types.ClusterDefinitionV5

	// Assert YAML definition version.
	// Order is important here! We try v5 first. Only if that fails, we try v4.
	switch def := definitionInterface.(type) {
	case *types.ClusterDefinitionV5:
		usesV5Definition = true
		defV5 = *def
	case *types.ClusterDefinitionV4:
		usesV4Definition = true
		defV4 = *def
	default:
		// Intentionally doing nothing.
	}

	// Check for wanted release from YAML definition.
	if usesV5Definition && defV5.ReleaseVersion != "" {
		wantedRelease = defV5.ReleaseVersion
	} else if usesV4Definition && defV4.ReleaseVersion != "" {
		wantedRelease = defV4.ReleaseVersion
	}

	// args overwrite definition content.
	if args.ReleaseVersion != "" {
		wantedRelease = args.ReleaseVersion
	}

	// If no release is set, use latest active release.
	// We need a release version here in order to be able to check capabilities.
	if wantedRelease == "" {
		latest, err := getLatestActiveReleaseVersion(clientWrapper, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
			fmt.Println(color.WhiteString("Determined release version %s is the latest, so this will be used.", latest))
		}
		wantedRelease = latest
	}

	// Fetch node pools capabilities info.
	capabilityService, err := capabilities.New(config.Config.Provider, clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
		fmt.Println(color.WhiteString("Fetching installation capabilities"))
	}

	nodePoolsEnabled, _ = capabilityService.HasCapability(wantedRelease, capabilities.NodePools)
	haMastersEnabled, _ = capabilityService.HasCapability(wantedRelease, capabilities.HAMasters)

	// Fail for edge cases:
	// - User uses v5 definition, but the installation doesn't support node pools.
	// - User uses v5 definition, but the release version requires v4.
	// - User uses v4 definition, but the release version requires v5.
	if nodePoolsEnabled && usesV4Definition {
		return nil, microerror.Maskf(errors.IncompatibleSettingsError, "please use a v5 definition or specify a release version that allows v4")
	} else if !nodePoolsEnabled && usesV5Definition {
		return nil, microerror.Maskf(errors.IncompatibleSettingsError, "please use a v4 definition or specify a release version that allows v5")
	}

	result := &creationResult{}

	if definitionInterface != nil {
		if usesV5Definition {
			result.DefinitionV5 = &defV5
			nodePoolsEnabled = true
		} else if usesV4Definition {
			result.DefinitionV4 = &defV4
		} else {
			return nil, microerror.Maskf(errors.YAMLNotParseableError, "unclear whether v4 or v5 cluster should be created")
		}
	}

	if nodePoolsEnabled {
		if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
			fmt.Println(color.WhiteString("Using the v5 API to create a cluster with node pool support"))
		}

		if result.DefinitionV5 == nil {
			result.DefinitionV5 = &types.ClusterDefinitionV5{}
		}

		// Validate inputs and set defaults.
		err = validateHAMasters(haMastersEnabled, &args, result.DefinitionV5)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		updateDefinitionFromFlagsV5(result.DefinitionV5, definitionFromFlagsV5{
			clusterName:    args.ClusterName,
			releaseVersion: args.ReleaseVersion,
			owner:          args.Owner,
			isHAMaster:     args.MasterHA,
		})

		id, hasErrors, err := addClusterV5(result.DefinitionV5, args, clientWrapper, auxParams)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result.ID = id
		result.HasErrors = hasErrors

	} else {
		if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
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

func toBoolPtr(t bool) *bool {
	return &t
}

func validateHAMasters(featureEnabled bool, args *Arguments, v5Definition *types.ClusterDefinitionV5) error {
	{
		if v5Definition.MasterNodes == nil && args.MasterHA == nil {
			// User tries to use the 'master' field in a version that supports HA masters.
			if featureEnabled && v5Definition.Master != nil {
				fmt.Println(color.YellowString("The 'master' attribute is deprecated.\nPlease remove the 'master' attribute from your cluster definition and use the 'master_nodes' attribute instead."))
			}
		} else if v5Definition.Master != nil {
			// User is trying to provide both 'master' and master nodes fields at the same time.
			return microerror.Mask(mustProvideSingleMasterTypeError)
		}
	}

	{
		// HA master has been enabled by cluster definition.
		hasHAMaster := v5Definition.MasterNodes != nil && v5Definition.MasterNodes.HighAvailability
		// HA master has been enabled by command-line flag.
		hasHAMasterFromFlag := args.MasterHA != nil && *args.MasterHA
		if hasHAMaster || hasHAMasterFromFlag {
			// User tries to use HA masters without it being supported.
			if !featureEnabled {
				return microerror.Mask(haMastersNotSupportedError)
			}
		} else if featureEnabled && v5Definition.Master == nil {
			if args.MasterHA == nil && v5Definition.MasterNodes == nil {
				// Check if 'master' field is set before defaulting to HA master.
				if args.OutputFormat != formatting.OutputFormatJSON && args.Verbose {
					fmt.Println(color.WhiteString("Using master node high availability by default."))
				}
				// Default to true if it is supported and not provided other value.
				args.MasterHA = toBoolPtr(true)
			}
		}
	}

	return nil
}
