package commands

import (
	"errors"
	"fmt"
	"os"
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
		Use:     "clusters",
		Short:   "List clusters",
		Long:    `Prints a list of all clusters you have access to`,
		PreRunE: checkListClusters,
		Run:     listClusters,
	}
)

const (
	listClustersActivityName = "list-clusters"
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
	if output != "" {
		fmt.Println(output)
	}
}

// clustersTable returns a table of clusters the user has access to
func clustersTable() (string, error) {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	clientConfig := client.Configuration{
		Endpoint:  endpoint,
		Timeout:   3 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return "", microerror.Mask(couldNotCreateClientError)
	}
	authHeader := "giantswarm " + token

	clusters, apiResponse, err := apiClient.GetClusters(authHeader,
		requestIDHeader, listClustersActivityName, cmdLine)
	if err != nil {
		return "", APIError{err.Error(), *apiResponse}
	}

	if len(clusters) == 0 {
		return "", nil
	}
	// table headers
	output := []string{color.CyanString("ID") + "|" + color.CyanString("NAME") + "|" + color.CyanString("CREATED") + "|" + color.CyanString("ORGANIZATION")}

	// sort clusters by organization
	slice.Sort(clusters[:], func(i, j int) bool {
		return clusters[i].Owner < clusters[j].Id
	})

	for _, cluster := range clusters {
		created := util.ShortDate(util.ParseDate(cluster.CreateDate))
		output = append(output,
			cluster.Id+"|"+
				cluster.Name+"|"+
				created+"|"+
				cluster.Owner)
	}

	return columnize.SimpleFormat(output), nil
}
