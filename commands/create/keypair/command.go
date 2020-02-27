// Package keypair implements the 'create keypair' command.
package keypair

import (
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/util"
)

var (
	// Command performs the "create keypair" function,
	Command = &cobra.Command{
		Use:    "keypair",
		Short:  "Create key pair",
		Long:   `Creates a new key pair for a cluster`,
		PreRun: printValidation,
		Run:    printResult,
	}

	arguments Arguments
)

const (
	activityName = "add-keypair"
)

// Arguments struct to pass to our business function and
// to the validation function
type Arguments struct {
	apiEndpoint              string
	authToken                string
	certificateOrganizations string
	clusterID                string
	commonNamePrefix         string
	description              string
	fileSystem               afero.Fs
	scheme                   string
	ttlHours                 int32
	userProvidedToken        string
	verbose                  bool
}

// collectArguments puts together arguments for our business function
// based on command line flags and config.
func collectArguments() (Arguments, error) {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	description := flags.Description
	if description == "" {
		description = "Added by user " + config.Config.Email + " using 'gsctl create keypair'"
	}

	ttl, err := util.ParseDuration(flags.TTL)
	if errors.IsInvalidDurationError(err) {
		return Arguments{}, microerror.Mask(errors.InvalidDurationError)
	} else if errors.IsDurationExceededError(err) {
		return Arguments{}, microerror.Mask(errors.DurationExceededError)
	} else if err != nil {
		return Arguments{}, microerror.Mask(errors.DurationExceededError)
	}

	return Arguments{
		apiEndpoint:              endpoint,
		authToken:                token,
		certificateOrganizations: flags.CertificateOrganizations,
		clusterID:                flags.ClusterID,
		commonNamePrefix:         flags.CNPrefix,
		description:              description,
		fileSystem:               config.FileSystem,
		scheme:                   scheme,
		ttlHours:                 int32(ttl.Hours()),
		userProvidedToken:        flags.Token,
		verbose:                  flags.Verbose,
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
	Command.Flags().StringVarP(&flags.ClusterID, "cluster", "c", "", "ID of the cluster to create a key pair for")
	Command.Flags().StringVarP(&flags.Description, "description", "d", "", "Description for the key pair")
	Command.Flags().StringVarP(&flags.CNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	Command.Flags().StringVarP(&flags.CertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	Command.Flags().StringVarP(&flags.TTL, "ttl", "", "30d", "Lifetime of the created key pair, e.g. 3h. Allowed units: h, d, w, m, y.")

	Command.MarkFlagRequired("cluster")
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	var argsErr error

	arguments, argsErr = collectArguments()
	if argsErr != nil {
		if errors.IsInvalidDurationError(argsErr) {
			fmt.Println(color.RedString("The value passed with --ttl is invalid."))
			fmt.Println("Please provide a number and a unit, e. g. '10h', '1d', '1w'.")
		} else if errors.IsDurationExceededError(argsErr) {
			fmt.Println(color.RedString("The expiration period passed with --ttl is too long."))
			fmt.Println("The maximum possible value is the eqivalent of 292 years.")
		} else {
			fmt.Println(color.RedString(argsErr.Error()))
		}
		os.Exit(1)
	}

	err := verifyPreconditions(arguments)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	headline := ""
	subtext := ""

	// TODO: handle specific errors
	switch {
	case err.Error() == "":
		return
	case errors.IsInvalidCNPrefixError(err):
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

func verifyPreconditions(args Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.clusterID == "" {
		return microerror.Mask(errors.ClusterNameOrIDMissingError)
	}

	// validate CN prefix character set
	if args.commonNamePrefix != "" {
		cnPrefixRE := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9@\\.-]*[a-zA-Z0-9]$")
		if !cnPrefixRE.MatchString(args.commonNamePrefix) {
			return microerror.Mask(errors.InvalidCNPrefixError)
		}
	}

	return nil
}

func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	result, err := createKeypair(arguments)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline string
		var subtext string

		switch {
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

	fmt.Println("Certificate and key files written to:")
	fmt.Println(result.caCertPath)
	fmt.Println(result.clientCertPath)
	fmt.Println(result.clientKeyPath)
}

// createKeypair is our business function talking to the API to create a keypair
// and return result or error
func createKeypair(args Arguments) (createKeypairResult, error) {
	result := createKeypairResult{
		apiEndpoint: args.apiEndpoint,
	}

	addKeyPairBody := &models.V4AddKeyPairRequest{
		Description:              &args.description,
		TTLHours:                 args.ttlHours,
		CnPrefix:                 args.commonNamePrefix,
		CertificateOrganizations: args.certificateOrganizations,
	}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return result, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = activityName

	response, err := clientWrapper.CreateKeyPair(args.clusterID, addKeyPairBody, auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clienterror.IsAccessForbiddenError(err) {
			return result, microerror.Mask(errors.AccessForbiddenError)
		}
		if clienterror.IsBadRequestError(err) {
			return result, microerror.Maskf(errors.BadRequestError, err.Error())
		}
		if clienterror.IsNotFoundError(err) {
			return result, microerror.Mask(errors.ClusterNotFoundError)
		}

		return result, microerror.Mask(err)
	}

	// success
	result.id = response.Payload.ID
	result.ttlHours = uint(response.Payload.TTLHours)

	// store credentials to file
	result.caCertPath = util.StoreCaCertificate(args.fileSystem, config.CertsDirPath,
		args.clusterID, response.Payload.CertificateAuthorityData)
	result.clientCertPath = util.StoreClientCertificate(args.fileSystem, config.CertsDirPath,
		args.clusterID, response.Payload.ID, response.Payload.ClientCertificateData)
	result.clientKeyPath = util.StoreClientKey(args.fileSystem, config.CertsDirPath,
		args.clusterID, response.Payload.ID, response.Payload.ClientKeyData)

	return result, nil
}
