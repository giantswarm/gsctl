package commands

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client/clienterror"
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
	scheme                   string
	authToken                string
	certificateOrganizations string
	clusterID                string
	commonNamePrefix         string
	description              string
	ttlHours                 int32
}

// function to create arguments based on command line flags and config
func defaultCreateKeypairArguments() (createKeypairArguments, error) {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	description := cmdDescription
	if description == "" {
		description = "Added by user " + config.Config.Email + " using 'gsctl create keypair'"
	}

	ttl, err := util.ParseDuration(cmdTTL)
	if IsInvalidDurationError(err) {
		return createKeypairArguments{}, microerror.Mask(invalidDurationError)
	} else if IsDurationExceededError(err) {
		return createKeypairArguments{}, microerror.Mask(durationExceededError)
	} else if err != nil {
		return createKeypairArguments{}, microerror.Mask(durationExceededError)
	}

	return createKeypairArguments{
		apiEndpoint:              endpoint,
		scheme:                   scheme,
		authToken:                token,
		certificateOrganizations: cmdCertificateOrganizations,
		clusterID:                cmdClusterID,
		commonNamePrefix:         cmdCNPrefix,
		description:              description,
		ttlHours:                 int32(ttl.Hours()),
	}, nil
}

type createKeypairResult struct {
	// cluster's API endpoint
	apiEndpoint string
	// path where we stored the CA file
	caCertPath string
	// path where we stored the client cert
	clientCertPath string
	// path where we stored the client's private key
	clientKeyPath string
	// key pair ID
	id string
	// TTL of the key pair in hours
	ttlHours uint
}

func init() {
	CreateKeypairCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key pair for")
	CreateKeypairCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key pair")
	CreateKeypairCommand.Flags().StringVarP(&cmdCNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	CreateKeypairCommand.Flags().StringVarP(&cmdCertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	CreateKeypairCommand.Flags().StringVarP(&cmdTTL, "ttl", "", "30d", "Lifetime of the created key pair, e.g. 3h. Allowed units: h, d, w, m, y.")

	CreateKeypairCommand.MarkFlagRequired("cluster")

	CreateCommand.AddCommand(CreateKeypairCommand)
}

func createKeyPairPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args, argsErr := defaultCreateKeypairArguments()
	if argsErr != nil {
		if IsInvalidDurationError(argsErr) {
			fmt.Println(color.RedString("The value passed with --ttl is invalid."))
			fmt.Println("Please provide a number and a unit, e. g. '10h', '1d', '1w'.")
		} else if IsDurationExceededError(argsErr) {
			fmt.Println(color.RedString("The expiration period passed with --ttl is too long."))
			fmt.Println("The maximum possible value is the eqivalent of 292 years.")
		} else {
			fmt.Println(color.RedString(argsErr.Error()))
		}
		os.Exit(1)
	}

	err := verifyCreateKeypairPreconditions(args)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	headline := ""
	subtext := ""

	// TODO: handle specific errors
	switch {
	case err.Error() == "":
		return
	case IsInvalidCNPrefixError(err):
		headline = "Bad characters in CN prefix (--cn-prefix)"
		subtext = "Please use these characters only: a-z A-Z 0-9 . @ -"
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

	// validate CN prefix character set
	if args.commonNamePrefix != "" {
		cnPrefixRE := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9@\\.-]*[a-zA-Z0-9]$")
		if !cnPrefixRE.MatchString(args.commonNamePrefix) {
			return microerror.Mask(invalidCNPrefixError)
		}
	}

	return nil
}

func createKeyPairRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args, _ := defaultCreateKeypairArguments()

	result, err := createKeypair(args)

	if err != nil {
		handleCommonErrors(err)

		var headline string
		var subtext string

		switch {
		case IsBadRequestError(err):
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
		util.Truncate(util.CleanKeypairID(result.id), 10, true),
		util.DurationPhrase(int(result.ttlHours)))
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

	addKeyPairBody := &models.V4AddKeyPairRequest{
		Description:              &args.description,
		TTLHours:                 args.ttlHours,
		CnPrefix:                 args.commonNamePrefix,
		CertificateOrganizations: args.certificateOrganizations,
	}
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = addKeyPairActivityName

	response, err := ClientV2.CreateKeyPair(args.clusterID, addKeyPairBody, auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return result, microerror.Mask(accessForbiddenError)
			} else if clientErr.HTTPStatusCode == http.StatusNotFound {
				return result, microerror.Mask(clusterNotFoundError)
			} else if clientErr.HTTPStatusCode == http.StatusForbidden {
				return result, microerror.Mask(accessForbiddenError)
			} else if clientErr.HTTPStatusCode == http.StatusBadRequest {
				return result, microerror.Maskf(badRequestError, clientErr.ErrorDetails)
			}
		}

		return result, microerror.Mask(err)
	}

	// success
	result.id = response.Payload.ID
	result.ttlHours = uint(response.Payload.TTLHours)

	// store credentials to file
	result.caCertPath = util.StoreCaCertificate(config.CertsDirPath,
		args.clusterID, response.Payload.CertificateAuthorityData)
	result.clientCertPath = util.StoreClientCertificate(config.CertsDirPath,
		args.clusterID, response.Payload.ID, response.Payload.ClientCertificateData)
	result.clientKeyPath = util.StoreClientKey(config.CertsDirPath,
		args.clusterID, response.Payload.ID, response.Payload.ClientKeyData)

	return result, nil
}
