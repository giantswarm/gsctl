package commands

import (
	"fmt"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsctl/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// ListEndpointsCommand performs the "list endpoints" function
	ListEndpointsCommand = &cobra.Command{
		Use:     "endpoints",
		Aliases: []string{"endpoint"},
		Short:   "List API endpoints",
		Long:    `Prints a list of API endpoints you have used so far`,
		Run:     listEndpoints,
	}
)

func init() {
	ListCommand.AddCommand(ListEndpointsCommand)
}

// listEndpoints prints a table with all endpoint URLs the user has used
func listEndpoints(cmd *cobra.Command, args []string) {
	output := endpointsTable()
	if output != "" {
		fmt.Println(output)
	}
}

// endpointsTable returns a table of clusters the user has access to
func endpointsTable() string {
	if len(config.Config.Endpoints) == 0 {
		return ""
	}

	// table headers
	output := []string{color.CyanString("ENDPOINT URL") + "|" + color.CyanString("SELECTED") + "|" + color.CyanString("LOGGED IN") + "|" + color.CyanString("EMAIL")}

	for endpoint := range config.Config.Endpoints {
		selected := "no"
		loggedIn := "no"
		email := "n/a"

		if endpoint == config.Config.SelectedEndpoint {
			selected = "yes"
		}

		if config.Config.Endpoints[endpoint].Token != "" && config.Config.Endpoints[endpoint].Email != "" {
			loggedIn = "yes"
			email = config.Config.Endpoints[endpoint].Email
		}

		output = append(output, endpoint+"|"+selected+"|"+loggedIn+"|"+email)
	}

	return columnize.SimpleFormat(output)
}
