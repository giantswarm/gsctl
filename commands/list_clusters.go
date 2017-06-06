package commands

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// ListClustersCommand performs the "list clusters" function
	ListClustersCommand = &cobra.Command{
		Use:     "clusters",
		Short:   "List clusters",
		Long:    `Prints a list of all clusters you have access to`,
		PreRunE: checkListClusters,
		Run:     listClusters,
	}
)

const (
	listClustersActivityName string = "list-clusters"
)

func init() {
	ListCommand.AddCommand(ListClustersCommand)
}

func checkListClusters(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	return nil
}

// listClusters prints a table with all clusters the user has access to
func listClusters(cmd *cobra.Command, args []string) {
	output, err := clustersTable()
	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		if _, ok := err.(APIError); ok {
			dumpAPIResponse((err).(APIError).APIResponse)
		}
		os.Exit(1)
	}
	fmt.Print(output)
}

// clustersTable returns a table of clusters the user has access to
func clustersTable() (string, error) {
	clientConfig := client.Configuration{
		Endpoint:  cmdAPIEndpoint,
		Timeout:   3 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)
	authHeader := "giantswarm " + config.Config.Token
	organizations, apiResponse, err := apiClient.GetUserOrganizations(authHeader, requestIDHeader, listClustersActivityName, cmdLine)
	if err != nil {
		gr, grErr := client.ParseGenericResponse(apiResponse.Payload)
		if grErr == nil {
			return "", fmt.Errorf("%s (Code: %s)", gr.Message, gr.Code)
		}
		return "", APIError{err.Error(), *apiResponse}
	}

	if apiResponse.Response.StatusCode == http.StatusOK {
		if len(organizations) == 0 {
			return "No organizations available", nil
		}

		// table headers
		output := []string{color.CyanString("ID") + "|" + color.CyanString("NAME") + "|" + color.CyanString("CREATED") + "|" + color.CyanString("ORGANIZATION")}

		// sort orgs by Id
		slice.Sort(organizations[:], func(i, j int) bool {
			return organizations[i].Id < organizations[j].Id
		})

		for _, org := range organizations {
			clustersResponse, _, err := apiClient.GetOrganizationClusters(authHeader, org.Id,
				requestIDHeader, listClustersActivityName, cmdLine)
			if err != nil {
				return "", APIError{err.Error(), *apiResponse}
			}

			for _, cluster := range clustersResponse.Data.Clusters {
				created := util.ShortDate(util.ParseDate(cluster.CreateDate))
				output = append(output,
					cluster.Id+"|"+
						cluster.Name+"|"+
						created+"|"+
						org.Id)
			}
		}
		return columnize.SimpleFormat(output), nil

	}

	return "", APIError{fmt.Sprintf("Unhandled response code: %v", apiResponse.Response.StatusCode), *apiResponse}
}
