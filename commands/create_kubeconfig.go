package commands

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// CreateKubeconfigCommand performs the "create kubeconfig" function
	CreateKubeconfigCommand = &cobra.Command{
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

  gsctl create kubeconfig -c my0c3 --self-contained ./kubeconfig.yaml

  gsctl create kubeconfig -c my0c3 --ttl 1 -d "Short lived key pair"

  gsctl create kubeconfig -c my0c3 --certificate-organizations system:masters
`,
		PreRun: createKubeconfigPreRunOutput,
		Run:    createKubeconfigRunOutput,
	}

	// cmdKubeconfigSelfContained is the command line flag for output of a
	// self-contained kubeconfig file
	cmdKubeconfigSelfContained = ""
)

const (
	createKubeconfigActivityName = "create-kubeconfig"
)

// createKubeconfigArguments is an argument struct to pass to our business
// function and to the validation function
type createKubeconfigArguments struct {
	apiEndpoint       string
	authToken         string
	clusterID         string
	description       string
	cnPrefix          string
	certOrgs          string
	ttlHours          int32
	selfContainedPath string
	force             bool
}

// defaultCreateKubeconfigArguments creates arguments based on command line
// flags and config and applies defaults
func defaultCreateKubeconfigArguments() createKubeconfigArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	description := cmdDescription
	if description == "" {
		description = "Added by user " + config.Config.Email + " using 'gsctl create kubeconfig'"
	}

	return createKubeconfigArguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		clusterID:         cmdClusterID,
		description:       description,
		cnPrefix:          cmdCNPrefix,
		certOrgs:          cmdCertificateOrganizations,
		ttlHours:          int32(cmdTTLDays) * 24,
		selfContainedPath: cmdKubeconfigSelfContained,
		force:             cmdForce,
	}
}

type createKubeconfigResult struct {
	// cluster's API endpoint
	apiEndpoint string
	// response body returned from a successful response
	keypairResponse gsclientgen.V4AddKeyPairResponse
	// path where we stored the CA file
	caCertPath string
	// path where we stored the client cert
	clientCertPath string
	// path where we stored the client's private key
	clientKeyPath string
	// absolute path for a self-contained kubeconfig file
	selfContainedPath string
}

// Kubeconfig is a struct used to create a kubectl configuration YAML file
type Kubeconfig struct {
	APIVersion     string                   `yaml:"apiVersion"`
	Kind           string                   `yaml:"kind"`
	Clusters       []KubeconfigNamedCluster `yaml:"clusters"`
	Users          []KubeconfigUser         `yaml:"users"`
	Contexts       []KubeconfigNamedContext `yaml:"contexts"`
	CurrentContext string                   `yaml:"current-context"`
	Preferences    struct{}                 `yaml:"preferences"`
}

// KubeconfigUser is a struct used to create a kubectl configuration YAML file
type KubeconfigUser struct {
	Name string                `yaml:"name"`
	User KubeconfigUserKeyPair `yaml:"user"`
}

// KubeconfigUserKeyPair is a struct used to create a kubectl configuration YAML file
type KubeconfigUserKeyPair struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

// KubeconfigNamedCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedCluster struct {
	Name    string            `yaml:"name"`
	Cluster KubeconfigCluster `yaml:"cluster"`
}

// KubeconfigCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
}

// KubeconfigNamedContext is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedContext struct {
	Name    string            `yaml:"name"`
	Context KubeconfigContext `yaml:"context"`
}

// KubeconfigContext is a struct used to create a kubectl configuration YAML file
type KubeconfigContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

func init() {
	CreateKubeconfigCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key pair")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdCNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdKubeconfigSelfContained, "self-contained", "", "", "Create a self-contained kubectl config with embedded credentials and write it to this path.")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdCertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	CreateKubeconfigCommand.Flags().BoolVarP(&cmdForce, "force", "", false, "If set, --self-contained will overwrite existing files without interactive confirmation.")
	CreateKubeconfigCommand.Flags().IntVarP(&cmdTTLDays, "ttl", "", 30, "Duration until expiry of the created key pair in days")

	CreateKubeconfigCommand.MarkFlagRequired("cluster")

	CreateCommand.AddCommand(CreateKubeconfigCommand)
}

// createKubeconfigPreRunOutput shows our pre-check results
func createKubeconfigPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultCreateKubeconfigArguments()
	err := verifyCreateKubeconfigPreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	headline := ""
	subtext := ""

	switch {
	case err.Error() == "":
		return
	case IsCommandAbortedError(err):
		headline = "File not overwritten, no kubeconfig created."
	case IsNotLoggedInError(err):
		headline = "You are not logged in."
		subtext = fmt.Sprintf("Use '%s login' to login or '--auth-token' to pass a valid auth token.", config.ProgramName)
	case IsKubectlMissingError(err):
		headline = "kubectl is not installed"
		if runtime.GOOS == "darwin" {
			subtext = "Please install via 'brew install kubernetes-cli' or visit\n"
			subtext += fmt.Sprintf("%s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "linux" {
			subtext = fmt.Sprintf("Please visit %s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "windows" {
			subtext = fmt.Sprintf("Please visit %s to download a recent kubectl binary.", kubectlWindowsInstallURL)
		}
	case IsClusterIDMissingError(err):
		headline = "No cluster specified"
		subtext = "Please use the --cluster or -c flag to indicate a cluster ID. Use --help for details."
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
func verifyCreateKubeconfigPreconditions(args createKubeconfigArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if args.apiEndpoint == "" {
		return microerror.Mask(endpointMissingError)
	}
	if args.clusterID == "" {
		return microerror.Mask(clusterIDMissingError)
	}

	kubectlOkay := util.CheckKubectl()
	if !kubectlOkay {
		return microerror.Mask(kubectlMissingError)
	}

	// ask for confirmation to overwrite existing file
	if args.selfContainedPath != "" && !args.force {
		if _, err := os.Stat(args.selfContainedPath); !os.IsNotExist(err) {
			confirmed := askForConfirmation("Do you want to overwrite " + args.selfContainedPath + " ?")
			if !confirmed {
				return microerror.Mask(commandAbortedError)
			}
		}
	}

	return nil
}

// createKubeconfig adds configuration for kubectl
func createKubeconfigRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultCreateKubeconfigArguments()
	result, err := createKubeconfig(args)

	if err != nil {

		headline := ""
		subtext := ""

		switch {
		case IsCouldNotCreateClientError(err):
			headline = "Failed to create API client."
			subtext = "Details: " + err.Error()
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
			subtext = "Context name to select is: giantswarm-" + args.clusterID
		case IsClusterNotFoundError(err):
			headline = fmt.Sprintf("Error: Cluster '%s' does not exist.", args.clusterID)
			subtext = "Please check the ID spelling or list clusters using 'gsctl list clusters'."
		case IsCouldNotWriteFileError(err):
			headline = "Error: File could not be written"
			subtext = fmt.Sprintf("Details: %s", err.Error())
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

	fmt.Println("Creating new key pair…")

	fmt.Printf("New key pair created with ID %s and expiry of %v hours\n",
		util.Truncate(util.CleanKeypairID(result.keypairResponse.Id), 10),
		result.keypairResponse.TtlHours)

	if result.selfContainedPath != "" {
		fmt.Printf("Self-contained kubectl config file written to: %s\n", result.selfContainedPath)

		fmt.Printf("\nTo make use of this file, run:\n\n")
		fmt.Println(color.YellowString("    export KUBECONFIG=" + result.selfContainedPath))
		fmt.Println(color.YellowString("    kubectl cluster-info\n"))

	} else {
		fmt.Println("Certificate and key files written to:")
		fmt.Println(result.caCertPath)
		fmt.Println(result.clientCertPath)
		fmt.Println(result.clientKeyPath)

		fmt.Printf("Switched to kubectl context 'giantswarm-%s'\n\n", args.clusterID)

		// final success message
		fmt.Println(color.GreenString("kubectl is set up. Check it using this command:\n"))
		fmt.Println(color.YellowString("    kubectl cluster-info\n"))
		fmt.Println(color.GreenString("Whenever you want to switch to using this context:\n"))
		fmt.Println(color.YellowString("    kubectl config use-context giantswarm-%s\n", args.clusterID))
	}
}

func createKubeconfig(args createKubeconfigArguments) (createKubeconfigResult, error) {
	result := createKubeconfigResult{}

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Maskf(couldNotCreateClientError, clientErr.Error())
	}

	authHeader := "giantswarm " + args.authToken

	// get cluster details
	clusterDetailsResponse, apiResponse, err := apiClient.GetCluster(
		authHeader,
		args.clusterID,
		requestIDHeader,
		createKubeconfigActivityName,
		cmdLine)
	if err != nil {
		return result, microerror.Maskf(err, fmt.Sprintf("HTTP status: %d", apiResponse.StatusCode))
	}

	result.apiEndpoint = clusterDetailsResponse.ApiEndpoint

	addKeyPairBody := gsclientgen.V4AddKeyPairBody{
		Description:              args.description,
		TtlHours:                 args.ttlHours,
		CnPrefix:                 cmdCNPrefix,
		CertificateOrganizations: cmdCertificateOrganizations,
	}

	clientConfig.Timeout = 60 * time.Second
	apiClient, clientErr = client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(couldNotCreateClientError)
	}

	keypairResponse, apiResponse, err := apiClient.AddKeyPair(authHeader,
		args.clusterID, addKeyPairBody, requestIDHeader,
		createKubeconfigActivityName, cmdLine)

	if err != nil {
		return result, microerror.Mask(err)
	}

	if apiResponse.StatusCode == 404 {
		// cluster not found
		return result, microerror.Mask(clusterNotFoundError)
	} else if apiResponse.StatusCode != 200 && apiResponse.StatusCode != 201 {
		return result, microerror.Maskf(unknownError,
			fmt.Sprintf("Unhandled response code: %v", apiResponse.StatusCode))
	}

	// success
	result.keypairResponse = *keypairResponse

	if args.selfContainedPath == "" {
		// modify the given kubeconfig file
		result.caCertPath = util.StoreCaCertificate(config.CertsDirPath,
			args.clusterID, keypairResponse.CertificateAuthorityData)
		result.clientCertPath = util.StoreClientCertificate(config.CertsDirPath,
			args.clusterID, keypairResponse.Id, keypairResponse.ClientCertificateData)
		result.clientKeyPath = util.StoreClientKey(config.CertsDirPath,
			args.clusterID, keypairResponse.Id, keypairResponse.ClientKeyData)

		// edit kubectl config
		if err := util.KubectlSetCluster(args.clusterID, clusterDetailsResponse.ApiEndpoint, result.caCertPath); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlClusterError)
		}
		if err := util.KubectlSetCredentials(args.clusterID, result.clientKeyPath, result.clientCertPath); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlCredentialsError)
		}
		if err := util.KubectlSetContext(args.clusterID); err != nil {
			return result, microerror.Mask(util.CouldNotSetKubectlContextError)
		}
		if err := util.KubectlUseContext(args.clusterID); err != nil {
			return result, microerror.Mask(util.CouldNotUseKubectlContextError)
		}
	} else {
		// create a self-contained kubeconfig
		kubeconfig := Kubeconfig{
			APIVersion:     "v1",
			Kind:           "Config",
			CurrentContext: "giantswarm-" + args.clusterID,
			Clusters: []KubeconfigNamedCluster{
				KubeconfigNamedCluster{
					Name: "giantswarm-" + args.clusterID,
					Cluster: KubeconfigCluster{
						Server: result.apiEndpoint,
						CertificateAuthorityData: base64.StdEncoding.EncodeToString([]byte(keypairResponse.CertificateAuthorityData)),
					},
				},
			},
			Contexts: []KubeconfigNamedContext{
				KubeconfigNamedContext{
					Name: "giantswarm-" + args.clusterID,
					Context: KubeconfigContext{
						Cluster: "giantswarm-" + args.clusterID,
						User:    "giantswarm-" + args.clusterID + "-user",
					},
				},
			},
			Users: []KubeconfigUser{
				KubeconfigUser{
					Name: "giantswarm-" + args.clusterID + "-user",
					User: KubeconfigUserKeyPair{
						ClientCertificateData: base64.StdEncoding.EncodeToString([]byte(keypairResponse.ClientCertificateData)),
						ClientKeyData:         base64.StdEncoding.EncodeToString([]byte(keypairResponse.ClientKeyData)),
					},
				},
			},
		}

		// create YAML
		yamlBytes, err := yaml.Marshal(&kubeconfig)
		if err != nil {
			return result, microerror.Mask(err)
		}

		err = ioutil.WriteFile(args.selfContainedPath, yamlBytes, 0600)
		if err != nil {
			return result, microerror.Maskf(couldNotWriteFileError, "could not write self-contained kubeconfig file")
		}

		result.selfContainedPath = args.selfContainedPath
	}

	return result, nil
}
