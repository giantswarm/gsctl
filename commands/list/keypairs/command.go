// Package keypairs implements the 'list keypairs' sub-command.
package keypairs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/util"
)

const (
	listKeypairsActivityName = "list-keypairs"

	outputFormatJSON  = "json"
	outputFormatTable = "table"

	outputJSONPrefix = ""
	outputJSONIndent = "  "
)

var (

	// Command performs the "list keypairs" function
	Command = &cobra.Command{
		Use:    "keypairs",
		Short:  "List key pairs for a cluster",
		Long:   `Prints a list of key pairs for a cluster`,
		PreRun: printValidation,
		Run:    printResult,
	}

	cmdOutput string

	arguments Arguments
)

// Arguments are the actual arguments used to call the
// listKeypairs() function.
type Arguments struct {
	apiEndpoint       string
	clusterNameOrID   string
	full              bool
	outputFormat      string
	token             string
	userProvidedToken string
	scheme            string
}

// collectArguments returns a new Arguments struct
// based on global variables (= command line options from cobra).
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		clusterNameOrID:   flags.ClusterID,
		full:              flags.Full,
		outputFormat:      cmdOutput,
		token:             token,
		userProvidedToken: flags.Token,
		scheme:            scheme,
	}
}

// listKeypairsResult is the data structure returned by the listKeypairs() function.
type listKeypairsResult struct {
	keypairs []*models.V4GetKeyPairsResponseItems
}

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().StringVarP(&flags.ClusterID, "cluster", "c", "", "Name/ID of the cluster to list key pairs for")
	Command.Flags().BoolVarP(&flags.Full, "full", "", false, "Enables output of full, untruncated values")
	Command.Flags().StringVarP(&cmdOutput, "output", "o", "table", "Use 'json' for JSON output. Defaults to human-friendly table output.")

	Command.MarkFlagRequired("cluster")
}

// printValidation does our pre-checks and shows errors, in case
// something is missing.
func printValidation(cmd *cobra.Command, extraArgs []string) {
	arguments = collectArguments()
	err := listKeypairsValidate(&arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}
}

// listKeypairsValidate validates our pre-conditions and returns an error in
// case something is missing.
// If no clusterNameOrID argument is given, and a default cluster can be determined,
// the Arguments given as argument will be modified to contain
// the clusterNameOrID field.
func listKeypairsValidate(args *Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.outputFormat != outputFormatJSON && args.outputFormat != outputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, fmt.Sprintf("Output format '%s' is unknown", args.outputFormat))
	}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return microerror.Mask(err)
	}

	if args.clusterNameOrID == "" {
		// use default cluster if possible
		clusterID, _ := clientWrapper.GetDefaultCluster(nil)
		if clusterID != "" {
			flags.ClusterID = clusterID
		} else {
			return microerror.Mask(errors.ClusterNameOrIDMissingError)
		}
	}

	return nil
}

// printResult is the function called to list keypairs and display
// errors in case they happen
func printResult(cmd *cobra.Command, extraArgs []string) {
	result, err := listKeypairs(arguments)

	// error output
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline string
		var subtext string

		switch {
		case errors.IsClusterNotFoundError(err):
			headline = "The cluster does not exist."
			subtext = fmt.Sprintf("We couldn't find the cluster '%s' via API endpoint %s.", arguments.clusterNameOrID, arguments.apiEndpoint)
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	if arguments.outputFormat == "json" {
		outputBytes, err := json.MarshalIndent(result.keypairs, outputJSONPrefix, outputJSONIndent)
		if err != nil {
			fmt.Println(color.RedString("Error while encoding JSON"))
			fmt.Printf("Details: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println(string(outputBytes))
	} else {
		// success output
		if len(result.keypairs) == 0 {
			fmt.Println(color.YellowString("No key pairs available for this cluster."))
			fmt.Println("You can create a new key pair using the 'gsctl create kubeconfig' or 'gsctl create keypair' command.")
		} else {
			output := []string{}

			headers := []string{
				color.CyanString("CREATED"),
				color.CyanString("EXPIRES"),
				color.CyanString("ID"),
				color.CyanString("DESCRIPTION"),
				color.CyanString("CN"),
				color.CyanString("O"),
			}
			output = append(output, strings.Join(headers, "|"))

			for _, keypair := range result.keypairs {
				createdTime := util.ParseDate(keypair.CreateDate)
				expiryTime := createdTime.Add(time.Duration(keypair.TTLHours) * time.Hour)
				expiryDuration := expiryTime.Sub(time.Now())
				expires := util.ShortDate(expiryTime)

				if expiryDuration < (24 * time.Hour) {
					expires = color.YellowString(expires)
				}

				// Idea: skip if expired, or only display when verbose
				row := []string{
					util.ShortDate(createdTime),
					expires,
					util.Truncate(formatting.CleanKeypairID(keypair.ID), 10, !arguments.full),
					keypair.Description,
					util.Truncate(keypair.CommonName, 24, !arguments.full),
					keypair.CertificateOrganizations,
				}
				output = append(output, strings.Join(row, "|"))
			}
			fmt.Println(columnize.SimpleFormat(output))
		}
	}
}

// listKeypairs fetches keypairs for a cluster from the API
// and returns them as a structured result.
func listKeypairs(args Arguments) (listKeypairsResult, error) {
	result := listKeypairsResult{}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return result, microerror.Mask(err)
	}

	clusterID, err := clustercache.GetID(args.apiEndpoint, args.clusterNameOrID, clientWrapper)
	if err != nil {
		return result, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listKeypairsActivityName

	response, err := clientWrapper.GetKeyPairs(clusterID, auxParams)
	if err != nil {
		if clienterror.IsUnauthorizedError(err) {
			return result, microerror.Mask(errors.NotAuthorizedError)
		}
		if clienterror.IsNotFoundError(err) {
			return result, microerror.Mask(errors.ClusterNotFoundError)
		}
		if clienterror.IsInternalServerError(err) {
			return result, microerror.Maskf(errors.InternalServerError, err.Error())
		}

		return result, microerror.Mask(err)
	}

	// sort key pairs by create date (descending)
	if len(response.Payload) > 1 {
		sort.Slice(response.Payload[:], func(i, j int) bool {
			return response.Payload[i].CreateDate < response.Payload[j].CreateDate
		})
	}

	result.keypairs = response.Payload

	return result, nil
}
