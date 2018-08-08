package commands

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client/clienterror"
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
	scheme      string
}

func defaultListClustersArguments() listClustersArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return listClustersArguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
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

		if clientErr, ok := err.(*clienterror.APIError); ok {
			fmt.Println(color.RedString(clientErr.ErrorMessage))
			if clientErr.ErrorDetails != "" {
				fmt.Println(clientErr.ErrorDetails)
			}
		} else {
			fmt.Println(color.RedString("Error: %s", err.Error()))
		}
		os.Exit(1)
	}

	if output != "" {
		fmt.Println(output)
	}
}

// clustersTable returns a table of clusters the user has access to
func clustersTable(args listClustersArguments) (string, error) {
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = listClustersActivityName

	response, err := ClientV2.GetClusters(auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			switch clientErr.HTTPStatusCode {
			case http.StatusUnauthorized:
				return "", microerror.Mask(notAuthorizedError)
			case http.StatusForbidden:
				return "", microerror.Mask(accessForbiddenError)
			}
		}

		return "", microerror.Mask(err)
	}

	if len(response.Payload) == 0 {
		return color.YellowString("No clusters available"), nil
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
	slice.Sort(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	for _, cluster := range response.Payload {
		created := util.ShortDate(util.ParseDate(cluster.CreateDate))
		releaseVersion := cluster.ReleaseVersion
		if releaseVersion == "" {
			releaseVersion = "n/a"
		}

		output = append(output, strings.Join([]string{
			cluster.ID,
			cluster.Owner,
			cluster.Name,
			releaseVersion,
			created,
		}, "|"))
	}

	return columnize.SimpleFormat(output), nil
}
