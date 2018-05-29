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
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// ListClustersCommand performs the "list clusters" function
	ListClustersCommand = &cobra.Command{
		Use:    "clusters",
		Short:  "List clusters",
		Long:   `Prints a list of all clusters you have access to`,
		PreRun: listClusterPreRunOutput,
		Run:    listClusterRunOutput,
	}
)

const (
	listClustersActivityName = "list-clusters"
)

type listClustersArguments struct {
	apiEndpoint string
	authToken   string
}

func defaultListClustersArguments() listClustersArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	return listClustersArguments{
		apiEndpoint: endpoint,
		authToken:   token,
	}
}

func init() {
	ListCommand.AddCommand(ListClustersCommand)
}

func listClusterPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultListClustersArguments()
	err := verifyListClusterPreconditions(args)

	if err == nil {
		return
	}

	handleCommonErrors(err)
}

func verifyListClusterPreconditions(args listClustersArguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if args.apiEndpoint == "" {
		return microerror.Mask(endpointMissingError)
	}

	return nil
}

// listClusters prints a table with all clusters the user has access to
func listClusterRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultListClustersArguments()

	output, err := clustersTable(args)
	if err != nil {
		handleCommonErrors(err)

		fmt.Println(color.RedString("Error: %s", err))
		if _, ok := err.(APIError); ok {
			dumpAPIResponse((err).(APIError).APIResponse)
		}
		os.Exit(1)
	}

	if output != "" {
		fmt.Println(output)
	}
}

// clustersTable returns a table of clusters the user has access to
func clustersTable(args listClustersArguments) (string, error) {
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   5 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return "", microerror.Mask(couldNotCreateClientError)
	}
	authHeader := "giantswarm " + args.authToken

	clusters, apiResponse, err := apiClient.GetClusters(authHeader,
		requestIDHeader, listClustersActivityName, cmdLine)
	if err != nil {
		if apiResponse.Response != nil && apiResponse.Response.StatusCode == http.StatusForbidden {
			return "", microerror.Mask(accessForbiddenError)
		}
		return "", APIError{err.Error(), *apiResponse}
	}

	if len(clusters) == 0 {
		return "", nil
	}
	// table headers
	output := []string{strings.Join([]string{
		color.CyanString("ID"),
		color.CyanString("ORGANIZATION"),
		color.CyanString("NAME"),
		color.CyanString("RELEASE"),
		color.CyanString("CREATED"),
	}, "|")}

	// sort clusters by ID
	slice.Sort(clusters[:], func(i, j int) bool {
		return clusters[i].Id < clusters[j].Id
	})

	for _, cluster := range clusters {
		created := util.ShortDate(util.ParseDate(cluster.CreateDate))
		releaseVersion := cluster.ReleaseVersion
		if releaseVersion == "" {
			releaseVersion = "n/a"
		}

		output = append(output, strings.Join([]string{
			cluster.Id,
			cluster.Owner,
			cluster.Name,
			releaseVersion,
			created,
		}, "|"))
	}

	return columnize.SimpleFormat(output), nil
}
