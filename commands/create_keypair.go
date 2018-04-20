package commands

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// CreateKeypairCommand performs the "create keypair" function
	CreateKeypairCommand = &cobra.Command{
		Use:    "keypair",
		Short:  "Create key pair",
		Long:   `Creates a new key pair for a cluster`,
		PreRun: createKeyPairPreRunOutput,
		Run:    createKeyPairRunOutput,
	}
)

const (
	addKeyPairActivityName = "add-keypair"
)

// argument struct to pass to our business function and
// to the validation function
type createKeypairArguments struct {
	apiEndpoint              string
	authToken                string
	certificateOrganizations string
	clusterID                string
	commonNamePrefix         string
	description              string
	ttlHours                 int32
}

// function to create arguments based on command line flags and config
func defaultCreateKeypairArguments() createKeypairArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	description := cmdDescription
	if description == "" {
		description = "Added by user " + config.Config.Email + " using 'gsctl create keypair'"
	}

	return createKeypairArguments{
		apiEndpoint:              endpoint,
		authToken:                token,
		certificateOrganizations: cmdCertificateOrganizations,
		clusterID:                cmdClusterID,
		commonNamePrefix:         cmdCNPrefix,
		description:              description,
		ttlHours:                 int32(cmdTTL) * 24,
	}
}

type createKeypairResult struct {
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
}

func init() {
	CreateKeypairCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key pair for")
	CreateKeypairCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key pair")
	CreateKeypairCommand.Flags().StringVarP(&cmdCNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	CreateKeypairCommand.Flags().StringVarP(&cmdCertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	CreateKeypairCommand.Flags().StringVarP(&cmdTTL, "ttl", "", "30d", "Lifetime of the created key pair, e.g. 3h. Default: 30d (30 days).")

	CreateKeypairCommand.MarkFlagRequired("cluster")

	CreateCommand.AddCommand(CreateKeypairCommand)
}

func createKeyPairPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultCreateKeypairArguments()
	err := verifyCreateKeypairPreconditions(args)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	headline := ""
	subtext := ""

	switch {
	case err.Error() == "":
		return
	default:
		headline = err.Error()
	}

	// print error output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

func verifyCreateKeypairPreconditions(args createKeypairArguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if args.apiEndpoint == "" {
		return microerror.Mask(endpointMissingError)
	}
	if args.clusterID == "" {
		return microerror.Mask(clusterIDMissingError)
	}

	return nil
}

func createKeyPairRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultCreateKeypairArguments()
	result, err := createKeypair(args)

	if err != nil {
		handleCommonErrors(err)

		var headline string
		var subtext string

		switch {
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
		util.Truncate(util.CleanKeypairID(result.keypairResponse.Id), 10),
		util.DurationPhrase(int(result.keypairResponse.TtlHours)))
	fmt.Println(color.GreenString(msg))

	fmt.Println("Certificate and key files written to:")
	fmt.Println(result.caCertPath)
	fmt.Println(result.clientCertPath)
	fmt.Println(result.clientKeyPath)
}

// createKeypair is our business function talking to the API to create a keypair
// and return result or error
func createKeypair(args createKeypairArguments) (createKeypairResult, error) {
	result := createKeypairResult{
		apiEndpoint: args.apiEndpoint,
	}

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   60 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Maskf(couldNotCreateClientError, clientErr.Error())
	}

	authHeader := "giantswarm " + args.authToken

	addKeyPairBody := gsclientgen.V4AddKeyPairBody{
		Description:              args.description,
		TtlHours:                 args.ttlHours,
		CnPrefix:                 args.commonNamePrefix,
		CertificateOrganizations: args.certificateOrganizations,
	}
	keypairResponse, apiResponse, err := apiClient.AddKeyPair(authHeader,
		args.clusterID, addKeyPairBody, requestIDHeader,
		addKeyPairActivityName, cmdLine)

	if err != nil {
		if apiResponse.Response != nil && apiResponse.Response.StatusCode == http.StatusForbidden {
			return result, microerror.Mask(accessForbiddenError)
		}
		return result, microerror.Mask(err)
	}

	if apiResponse.StatusCode == http.StatusNotFound {
		return result, microerror.Mask(clusterNotFoundError)
	} else if apiResponse.StatusCode == http.StatusForbidden {
		return result, microerror.Mask(accessForbiddenError)
	} else if apiResponse.StatusCode != http.StatusOK && apiResponse.StatusCode != 201 {
		return result, microerror.Maskf(unknownError,
			"Unhandled response code: %v", apiResponse.StatusCode)
	}

	// success
	result.keypairResponse = *keypairResponse

	// store credentials to file
	result.caCertPath = util.StoreCaCertificate(config.CertsDirPath,
		args.clusterID, keypairResponse.CertificateAuthorityData)
	result.clientCertPath = util.StoreClientCertificate(config.CertsDirPath,
		args.clusterID, keypairResponse.Id, keypairResponse.ClientCertificateData)
	result.clientKeyPath = util.StoreClientKey(config.CertsDirPath,
		args.clusterID, keypairResponse.Id, keypairResponse.ClientKeyData)

	return result, nil
}
