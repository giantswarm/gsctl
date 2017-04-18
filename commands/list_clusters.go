package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
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

// list all clusters the user has access to
func listClusters(cmd *cobra.Command, args []string) {
	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	authHeader := "giantswarm " + config.Config.Token
	orgsResponse, apiResponse, err := client.GetUserOrganizations(authHeader, requestIDHeader, listClustersActivityName, cmdLine)
	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}
	if orgsResponse.StatusCode == apischema.STATUS_CODE_DATA {
		var organizations = orgsResponse.Data
		if len(organizations) == 0 {
			fmt.Println(color.YellowString("No organizations available"))
		} else {
			sort.Strings(organizations)
			output := []string{color.CyanString("ID") + "|" + color.CyanString("NAME") + "|" + color.CyanString("CREATED") + "|" + color.CyanString("ORGANIZATION")}
			for _, orgName := range organizations {
				clustersResponse, _, err := client.GetOrganizationClusters(authHeader, orgName, requestIDHeader, listClustersActivityName, cmdLine)
				if err != nil {
					fmt.Println(color.RedString("Error: %s", err))
					dumpAPIResponse(*apiResponse)
					os.Exit(1)
				}
				for _, cluster := range clustersResponse.Data.Clusters {
					created := util.ShortDate(util.ParseDate(cluster.CreateDate))
					output = append(output,
						cluster.Id+"|"+
							cluster.Name+"|"+
							created+"|"+
							orgName)
				}
			}
			fmt.Println(columnize.SimpleFormat(output))
		}
	} else {
		fmt.Println(color.RedString("Unhandled response code: %v", orgsResponse.StatusCode))
		dumpAPIResponse(*apiResponse)
	}
}
