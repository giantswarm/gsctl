package commands

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

const (
	listKeypairsActivityName = "list-keypairs"
)

var (

	// ListKeypairsCommand performs the "list keypairs" function
	ListKeypairsCommand = &cobra.Command{
		Use:    "keypairs",
		Short:  "List key pairs for a cluster",
		Long:   `Prints a list of key pairs for a cluster`,
		PreRun: listKeypairsValidationOutput,
		Run:    listKeypairsOutput,
	}
)

// listKeypairsArguments are the actual arguments used to call the
// listKeypairs() function.
type listKeypairsArguments struct {
	apiEndpoint string
	clusterID   string
	full        bool
	token       string
	scheme      string
}

// defaultListKeypairsArguments returns a new listKeypairsArguments struct
// based on global variables (= command line options from cobra).
func defaultListKeypairsArguments() listKeypairsArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return listKeypairsArguments{
		apiEndpoint: endpoint,
		clusterID:   cmdClusterID,
		full:        cmdFull,
		token:       token,
		scheme:      scheme,
	}
}

// listKeypairsResult is the data structure returned by the listKeypairs() function.
type listKeypairsResult struct {
	keypairs []gsclientgen.KeyPairModel
}

func init() {
	ListKeypairsCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to list key pairs for")
	ListKeypairsCommand.Flags().BoolVarP(&cmdFull, "full", "", false, "Enables output of full, untruncated values")

	ListKeypairsCommand.MarkFlagRequired("cluster")
	ListCommand.AddCommand(ListKeypairsCommand)
}

// listKeypairsValidationOutput does our pre-checks and shows errors, in case
// something is missing.
func listKeypairsValidationOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListKeypairsArguments()
	err := listKeypairsValidate(&args)
	if err != nil {
		handleCommonErrors(err)

		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}
}

// listKeypairsValidate validates our pre-conditions and returns an error in
// case something is missing.
// If no clusterID argument is given, and a default cluster can be determined,
// the listKeypairsArguments given as argument will be modified to contain
// the clusterID field.
func listKeypairsValidate(args *listKeypairsArguments) error {
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}
	if args.clusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster(requestIDHeader, listKeypairsActivityName, cmdLine, cmdAPIEndpoint)
		if clusterID != "" {
			cmdClusterID = clusterID
		} else {
			return microerror.Mask(clusterIDMissingError)
		}
	}

	return nil
}

// listKeypairsOutput is the function called to list keypairs and display
// errors in case they happen
func listKeypairsOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListKeypairsArguments()
	result, err := listKeypairs(args)

	// error output
	if err != nil {
		handleCommonErrors(err)

		var headline string
		var subtext string

		switch {
		case IsClusterNotFoundError(err):
			headline = "The cluster does not exist."
			subtext = fmt.Sprintf("We couldn't find a cluster with the ID '%s' via API endpoint %s.", args.clusterID, args.apiEndpoint)
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

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
			created := util.ParseDate(keypair.CreateDate)
			expires := util.ParseDate(keypair.CreateDate).Add(time.Duration(keypair.TtlHours) * time.Hour)

			// Idea: skip if expired, or only display when verbose
			row := []string{
				util.ShortDate(created),
				util.ShortDate(expires),
				util.Truncate(util.CleanKeypairID(keypair.Id), 10, !args.full),
				keypair.Description,
				util.Truncate(keypair.CommonName, 24, !args.full),
				keypair.CertificateOrganizations,
			}
			output = append(output, strings.Join(row, "|"))
		}
		fmt.Println(columnize.SimpleFormat(output))
	}
}

// listKeypairs fetches keypairs for a cluster from the API
// and returns them as a structured result.
func listKeypairs(args listKeypairsArguments) (listKeypairsResult, error) {
	result := listKeypairsResult{}

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   20 * time.Second,
		UserAgent: config.UserAgent(),
	}

	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(couldNotCreateClientError)
	}
	authHeader := args.scheme + " " + args.token
	keypairsResponse, apiResponse, err := apiClient.GetKeyPairs(authHeader,
		cmdClusterID, requestIDHeader, listKeypairsActivityName, cmdLine)

	if err != nil {

		if apiResponse.StatusCode >= 500 {
			return result, microerror.Maskf(internalServerError, err.Error())
		} else if apiResponse.StatusCode == http.StatusNotFound {
			return result, microerror.Mask(clusterNotFoundError)
		} else if apiResponse.StatusCode == http.StatusUnauthorized {
			return result, microerror.Mask(notAuthorizedError)
		}
		return result, microerror.Mask(err)
	}

	if apiResponse.StatusCode != http.StatusOK {
		return result, microerror.Mask(unknownError)
	}

	// sort key pairs by create date (descending)
	if len(keypairsResponse) > 1 {
		slice.Sort(keypairsResponse[:], func(i, j int) bool {
			return keypairsResponse[i].CreateDate < keypairsResponse[j].CreateDate
		})
	}

	result.keypairs = keypairsResponse

	return result, nil
}
