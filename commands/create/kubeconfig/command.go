// Package kubeconfig implements the 'create kubeconfig' command.
package kubeconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/client/key_pairs"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/kubeconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/util"
)

var (
	// Command performs the "create kubeconfig" function
	Command = &cobra.Command{
		Use:   "kubeconfig",
		Short: "Configure kubectl",
		Long: `Creates or modifies kubectl configuration to access your Giant Swarm
Kubernetes cluster

By executing this command, you create a new key pair for your cluster to
authenticate with as a kubectl user.

By default, your kubectl config is modified to add user, cluster, and context
entries. The config file is assumed to be in $HOME/.kube/config. If set, the
path from the $KUBECONFIG environment variable is used. Certificate files are
stored in the "certs" subfolder of the gsctl config directory. See 'gsctl info'.

Alternatively, the --self-contained <path> flag can be used to create a new
config file with included certificates.

Examples:

  gsctl create kubeconfig -c my0c3

  gsctl create kubeconfig -c "Production cluster" --self-contained ./kubeconfig.yaml

  gsctl create kubeconfig -c my0c3 --ttl 3h -d "Key pair living for only 3 hours"

  gsctl create kubeconfig -c "Development cluster" --certificate-organizations system:masters
`,
		PreRun: createKubeconfigPreRunOutput,
		Run:    createKubeconfigRunOutput,
	}

	// cmdKubeconfigSelfContained is the command line flag for output of a
	// self-contained kubeconfig file
	cmdKubeconfigSelfContained = ""

	// flag for setting a kubectl context name to use
	cmdKubeconfigContextName = ""

	arguments Arguments
)

const (
	createKubeconfigActivityName = "create-kubeconfig"

	// url to intallation instructions
	kubectlInstallURL = "https://kubernetes.io/docs/tasks/tools/install-kubectl/"

	// windows download page
	kubectlWindowsInstallURL = "https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md"

	// tenant internal api prefix
	tenantInternalAPIPrefix = "internal-api"

	urlDelimiter = "."

	// Maximum safe TTL (in hours)
	maxSafeTTLHours = 30 * 24 // 30 days
)

// Arguments is an argument struct to pass to our business
// function and to the validation function
type Arguments struct {
	apiEndpoint       string
	authToken         string
	certOrgs          string
	clusterNameOrID   string
	cnPrefix          string
	contextName       string
	description       string
	fileSystem        afero.Fs
	force             bool
	tenantInternal    bool
	outputFormat      string
	scheme            string
	selfContainedPath string
	ttlHours          int32
	userProvidedToken string
	verbose           bool
}

// collectArguments gathers arguments based on command line
// flags and config and applies defaults.
func collectArguments() (Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	description := flags.Description
	if description == "" {
		description = "Added by user " + config.Config.Email + " using 'gsctl create kubeconfig'"
	}

	contextName := cmdKubeconfigContextName

	ttl, err := util.ParseDuration(flags.TTL)
	if errors.IsInvalidDurationError(err) {
		return Arguments{}, microerror.Mask(errors.InvalidDurationError)
	} else if errors.IsDurationExceededError(err) {
		return Arguments{}, microerror.Mask(errors.DurationExceededError)
	} else if err != nil {
		return Arguments{}, microerror.Mask(err)
	}

	// hack..
	// cobra sets defaults from other commands to the OutputFormat flag
	// but we don't have "table" here, so if it's "table", set it to empty string
	if flags.OutputFormat == "table" {
		flags.OutputFormat = ""
	}

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		certOrgs:          flags.CertificateOrganizations,
		clusterNameOrID:   flags.ClusterID,
		cnPrefix:          flags.CNPrefix,
		contextName:       contextName,
		description:       description,
		fileSystem:        config.FileSystem,
		force:             flags.Force,
		tenantInternal:    flags.TenantInternal,
		outputFormat:      flags.OutputFormat,
		scheme:            scheme,
		selfContainedPath: cmdKubeconfigSelfContained,
		ttlHours:          int32(ttl.Hours()),
		userProvidedToken: flags.Token,
		verbose:           flags.Verbose,
	}, nil
}

type createKubeconfigResult struct {
	// cluster's API endpoint
	apiEndpoint string
	// path where we stored the CA file
	caCertPath string
	// path where we stored the client cert
	clientCertPath string
	// path where we stored the client's private key
	clientKeyPath string
	// absolute path for a self-contained kubeconfig file
	selfContainedPath string
	// kubeconfig yaml bytes
	selfContainedYAMLBytes []byte
	// the context name applied
	contextName string
	// key pair ID
	id string
	// TTL of the key pair in hours
	ttlHours uint
}

// JSONOutput contains the fields included in JSON output of the create kubeconfig command when called with json output flag
type JSONOutput struct {
	// Result of the command. should be 'ok'
	Result string `json:"result"`
	// KubeConfig is a string containing the kubeconfig
	KubeConfig string `json:"kubeconfig,omitempty"`
	// Error which occured
	Error error `json:"error,omitempty"`
}

func init() {
	Command.Flags().StringVarP(&flags.ClusterID, "cluster", "c", "", "Name or ID of the cluster")
	Command.Flags().StringVarP(&flags.Description, "description", "d", "", "Description for the key pair")
	Command.Flags().StringVarP(&flags.CNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	Command.Flags().StringVarP(&cmdKubeconfigSelfContained, "self-contained", "", "", "Create a self-contained kubectl config with embedded credentials and write it to this path.")
	Command.Flags().StringVarP(&cmdKubeconfigContextName, "context", "", "", "Set a custom context name. Defaults to 'giantswarm-<cluster-id>'.")
	Command.Flags().StringVarP(&flags.CertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	Command.Flags().BoolVarP(&flags.Force, "force", "", false, "If set, --self-contained will overwrite existing files without interactive confirmation. Also, there will not be any confirmation for TTL > 30d.")
	Command.Flags().BoolVarP(&flags.TenantInternal, "tenant-internal", "", false, "If set, kubeconfig will be rendered with internal Kubernets API address.")
	Command.Flags().StringVarP(&flags.TTL, "ttl", "", "1d", "Lifetime of the created key pair, e.g. 3h. Allowed units: h, d, w, m, y.")
	Command.Flags().StringVarP(&flags.OutputFormat, "output", "", "", fmt.Sprintf("Output format. Specifying '%s' will change output to be JSON formatted.", formatting.OutputFormatJSON))

	Command.MarkFlagRequired("cluster")
}

// createKubeconfigPreRunOutput shows our pre-check results
func createKubeconfigPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	var argsErr error

	arguments, argsErr = collectArguments()
	if argsErr != nil {
		if errors.IsInvalidDurationError(argsErr) {
			fmt.Println(color.RedString("The value passed with --ttl is invalid."))
			fmt.Println("Please provide a number and a unit, e. g. '10h', '1d', '1w'.")
		} else if errors.IsDurationExceededError(argsErr) {
			fmt.Println(color.RedString("The expiration period passed with --ttl is too long."))
			fmt.Println("The maximum possible value is the equivalent of 292 years.")
		} else {
			fmt.Println(color.RedString(argsErr.Error()))
		}
		os.Exit(1)
	}

	if !arguments.force && arguments.ttlHours >= maxSafeTTLHours {
		fmt.Println("The desired expiry date is pretty far away.")
		fmt.Println("There is no way to revoke keypairs once they've been created.")
		question := fmt.Sprintf("Are you sure you want to set the TTL to %s?", flags.TTL)
		confirmed := confirm.Ask(question)
		if !confirmed {
			os.Exit(0)
		}
	}

	err := verifyCreateKubeconfigPreconditions(arguments, cmdLineArgs)
	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	var headline string
	var subtext string

	switch {
	case errors.IsCommandAbortedError(err):
		headline = "File not overwritten, no kubeconfig created."
	case errors.IsKubectlMissingError(err):
		headline = "kubectl is not installed"
		if runtime.GOOS == "darwin" {
			subtext = "Please install via 'brew install kubernetes-cli' or visit\n"
			subtext += fmt.Sprintf("%s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "linux" {
			subtext = fmt.Sprintf("Please visit %s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "windows" {
			subtext = fmt.Sprintf("Please visit %s to download a recent kubectl binary.", kubectlWindowsInstallURL)
		}
	case errors.IsInvalidCNPrefixError(err):
		headline = "Bad characters in CN prefix (--cn-prefix)"
		subtext = "Please use these characters only: a-z A-Z 0-9 . @ -"
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

// verifyCreateKubeconfigPreconditions checks if all preconditions are met and
// returns nil if yes, error if not
func verifyCreateKubeconfigPreconditions(args Arguments, cmdLineArgs []string) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.clusterNameOrID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	}
	if args.outputFormat != "" && args.outputFormat != formatting.OutputFormatJSON {
		return microerror.Maskf(errors.OutputFormatInvalidError, fmt.Sprintf("Output format '%s' is is invalid for gsctl create kubeconfig. Valid options: '%s'", args.outputFormat, formatting.OutputFormatJSON))
	}

	// validate CN prefix character set
	if args.cnPrefix != "" {
		cnPrefixRE := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9@\\.-]*[a-zA-Z0-9]$")
		if !cnPrefixRE.MatchString(args.cnPrefix) {
			return microerror.Mask(errors.InvalidCNPrefixError)
		}
	}

	kubectlOkay := util.CheckKubectl()
	if !kubectlOkay {
		return microerror.Mask(errors.KubectlMissingError)
	}

	// ask for confirmation to overwrite existing file
	if args.selfContainedPath != "" && !args.force {
		if _, err := os.Stat(args.selfContainedPath); !os.IsNotExist(err) {
			confirmed := confirm.Ask("Do you want to overwrite " + args.selfContainedPath + " ?")
			if !confirmed {
				return microerror.Mask(errors.CommandAbortedError)
			}
		}
	}

	return nil
}

// createKubeconfig adds configuration for kubectl
func createKubeconfigRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	ctx := context.Background()

	result, err := createKubeconfig(ctx, arguments)

	if arguments.outputFormat == formatting.OutputFormatJSON {
		printJSONOutput(result, err)
		return
	}

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline string
		var subtext string

		switch {
		case util.IsCouldNotSetKubectlClusterError(err):
			headline = "Error: " + err.Error()
			subtext = "API endpoint would be: " + result.apiEndpoint
			subtext += "\nCA file path would be: " + result.caCertPath
		case util.IsCouldNotSetKubectlCredentialsError(err):
			headline = "Error: " + err.Error()
			subtext = "Client key path would be: " + result.clientKeyPath
			subtext += "\nClient certificate path would be: " + result.clientCertPath
		case util.IsCouldNotSetKubectlContextError(err):
			headline = "Error: " + err.Error()
		case util.IsCouldNotUseKubectlContextError(err):
			headline = "Error: " + err.Error()
			subtext = "Context name to select is: giantswarm-" + arguments.clusterNameOrID
		case errors.IsClusterNotFoundError(err):
			headline = fmt.Sprintf("Error: Cluster '%s' does not exist.", arguments.clusterNameOrID)
			subtext = "Please check the name/ID spelling or list clusters using 'gsctl list clusters'."
		case errors.IsCouldNotWriteFileError(err):
			headline = "Error: File could not be written"
			subtext = fmt.Sprintf("Details: %s", err.Error())
		case errors.IsBadRequestError(err):
			headline = "API Error 400: Bad Request"
			subtext = "The key pair could not be created with the given parameters. Please try a shorter expiry period (--ttl)\n"
			subtext += "and check the other arguments, too. Please contact the Giant Swarm support team if you need assistance."
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

	// Success output

	msg := fmt.Sprintf("New key pair created with ID %s and expiry of %v",
		util.Truncate(formatting.CleanKeypairID(result.id), 10, true),
		util.DurationPhrase(int(result.ttlHours)))
	fmt.Println(color.GreenString(msg))

	if result.selfContainedPath != "" {
		fmt.Printf("Self-contained kubectl config file written to: %s\n", result.selfContainedPath)

		fmt.Printf("\nTo make use of this file, run:\n\n")
		fmt.Println(color.YellowString("    export KUBECONFIG=" + result.selfContainedPath))
		fmt.Println(color.YellowString("    kubectl cluster-info\n"))

	} else {
		if arguments.verbose {
			fmt.Println(color.WhiteString("Certificate and key files written to:"))
			fmt.Println(color.WhiteString(result.caCertPath))
			fmt.Println(color.WhiteString(result.clientCertPath))
			fmt.Println(color.WhiteString(result.clientKeyPath))
		}

		fmt.Printf("Switched to kubectl context '%s'\n\n", result.contextName)

		// final success message
		fmt.Println(color.GreenString("kubectl is set up. Check it using this command:\n"))
		fmt.Println(color.YellowString("    kubectl cluster-info\n"))
		fmt.Println(color.GreenString("Whenever you want to switch to using this context:\n"))
		fmt.Println(color.YellowString("    kubectl config use-context %s\n", result.contextName))
	}
}

func printJSONOutput(result createKubeconfigResult, creationErr error) {
	var outputBytes []byte
	var err error
	var jsonResult JSONOutput

	// handle errors
	if creationErr != nil {
		jsonResult = JSONOutput{Result: "error", Error: creationErr}
	} else {
		jsonResult = JSONOutput{Result: "ok", KubeConfig: string(result.selfContainedYAMLBytes)}
	}

	outputBytes, err = json.MarshalIndent(jsonResult, formatting.OutputJSONPrefix, formatting.OutputJSONIndent)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(outputBytes))
	if creationErr != nil {
		os.Exit(1)
	}
}

// getClusterDetails fetches cluster details to get the tenant cluster API endpoint,
// and attempts first v5 and then falls back to v4.
func getClusterDetails(clientWrapper *client.Wrapper, clusterID string, auxParams *client.AuxiliaryParams, verbose bool) (string, error) {
	// Try v5 first, then fall back to v4.
	if verbose {
		fmt.Println(color.WhiteString("Fetching cluster details using the v5 API endpoint"))
	}
	clusterDetailsResponseV5, err := clientWrapper.GetClusterV5(clusterID, auxParams)
	if err == nil {
		return clusterDetailsResponseV5.Payload.APIEndpoint, nil
	}

	if clienterror.IsNotFoundError(err) || clienterror.IsBadRequestError(err) {
		// If v5 failed with a 404 Not Found or 400 Bad Request error, we try v4.
		if verbose {
			fmt.Println(color.WhiteString("Cluster not found via the v5 endpoint. Attempting v4 endpoint."))
		}
		clusterDetailsResponseV4, err := clientWrapper.GetClusterV4(clusterID, auxParams)
		if err == nil {
			return clusterDetailsResponseV4.Payload.APIEndpoint, nil
		}

		if clientErr, ok := err.(*clienterror.APIError); ok {
			apiErr := &microerror.Error{
				Kind: clientErr.ErrorMessage,
			}

			return "", microerror.Maskf(apiErr, "HTTP Status: %d, %s", clientErr.HTTPStatusCode, clientErr.ErrorMessage)
		}

		return "", microerror.Mask(err)
	}

	return "", microerror.Mask(err)
}

// createKubeconfig is our business function talking to the API to create a keypair
// and creating a new kubectl context
func createKubeconfig(ctx context.Context, args Arguments) (createKubeconfigResult, error) {
	result := createKubeconfigResult{}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return result, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.apiEndpoint, args.clusterNameOrID, clientWrapper)
	if err != nil {
		return result, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = createKubeconfigActivityName

	result.apiEndpoint, err = getClusterDetails(clientWrapper, clusterID, auxParams, args.verbose)
	if err != nil {
		return createKubeconfigResult{}, microerror.Mask(err)
	}

	// Set internal API endpoint if requested.
	if args.tenantInternal {
		baseEndpoint := strings.Split(result.apiEndpoint, urlDelimiter)[1:]
		result.apiEndpoint = fmt.Sprintf("https://%s.%s", tenantInternalAPIPrefix, strings.Join(baseEndpoint, urlDelimiter))
	}

	addKeyPairBody := &models.V4AddKeyPairRequest{
		Description:              &args.description,
		TTLHours:                 args.ttlHours,
		CnPrefix:                 args.cnPrefix,
		CertificateOrganizations: args.certOrgs,
	}

	response, err := clientWrapper.CreateKeyPair(clusterID, addKeyPairBody, auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clienterror.IsAccessForbiddenError(err) {
			return result, microerror.Mask(errors.AccessForbiddenError)
		}
		if clienterror.IsNotFoundError(err) {
			return result, microerror.Mask(errors.ClusterNotFoundError)
		}
		if clienterror.IsBadRequestError(err) {
			return result, microerror.Maskf(errors.BadRequestError, err.Error())
		}

		return result, microerror.Mask(err)
	}

	// success
	result.id = response.Payload.ID
	result.ttlHours = uint(response.Payload.TTLHours)

	if args.outputFormat == formatting.OutputFormatJSON {
		yamlBytes, err := createKubeconfigYAML(ctx, clusterID, result.apiEndpoint, response)
		if err != nil {
			return result, microerror.Mask(err)
		}

		result.selfContainedYAMLBytes = yamlBytes

	} else if args.selfContainedPath == "" {
		// modify the given kubeconfig file
		result.caCertPath = util.StoreCaCertificate(args.fileSystem, config.CertsDirPath,
			clusterID, response.Payload.CertificateAuthorityData)
		result.clientCertPath = util.StoreClientCertificate(args.fileSystem, config.CertsDirPath,
			clusterID, response.Payload.ID, response.Payload.ClientCertificateData)
		result.clientKeyPath = util.StoreClientKey(args.fileSystem, config.CertsDirPath,
			clusterID, response.Payload.ID, response.Payload.ClientKeyData)
		result.contextName = args.contextName
		if result.contextName == "" {
			result.contextName = "giantswarm-" + clusterID
		}

		// edit kubectl config
		if err := util.KubectlSetCluster(clusterID, result.apiEndpoint, result.caCertPath); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlClusterError)
		}
		if err := util.KubectlSetCredentials(clusterID, result.clientKeyPath, result.clientCertPath); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlCredentialsError)
		}
		if err := util.KubectlSetContext(result.contextName, clusterID); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlContextError)
		}
		if err := util.KubectlUseContext(result.contextName); err != nil {
			return result, microerror.Mask(util.CouldNotUseKubectlContextError)
		}
	} else {
		// create a self-contained kubeconfig
		yamlBytes, err := createKubeconfigYAML(ctx, clusterID, result.apiEndpoint, response)
		if err != nil {
			return result, microerror.Mask(err)
		}

		err = afero.WriteFile(args.fileSystem, args.selfContainedPath, yamlBytes, 0600)
		if err != nil {
			return result, microerror.Maskf(errors.CouldNotWriteFileError, "could not write self-contained kubeconfig file")
		}

		result.selfContainedPath = args.selfContainedPath
	}

	return result, nil
}

func createKubeconfigYAML(ctx context.Context, clusterID, apiEndpoint string, response *key_pairs.AddKeyPairOK) ([]byte, error) {
	var yamlBytes []byte
	logger, err := micrologger.New(micrologger.Config{
		IOWriter: new(bytes.Buffer), // to suppress any log output
	})
	if err != nil {
		return yamlBytes, microerror.Mask(err)
	}
	{
		c := k8srestconfig.Config{
			Logger: logger,

			Address:   apiEndpoint,
			InCluster: false,
			TLS: k8srestconfig.ConfigTLS{
				CAData:  []byte(response.Payload.CertificateAuthorityData),
				CrtData: []byte(response.Payload.ClientCertificateData),
				KeyData: []byte(response.Payload.ClientKeyData),
			},
		}
		restConfig, err := k8srestconfig.New(c)
		if err != nil {
			return yamlBytes, microerror.Mask(err)
		}

		yamlBytes, err = kubeconfig.NewKubeConfigForRESTConfig(ctx, restConfig, fmt.Sprintf("giantswarm-%s", clusterID), "")
		if err != nil {
			return yamlBytes, microerror.Mask(err)
		}

		return yamlBytes, nil
	}
}
