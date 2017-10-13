package commands

import (
	"fmt"
	"strings"

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
		return fmt.Sprintf("No endpoints configured.\n\nTo add an endpoint and authenticate for it, use\n\n\t%s\n",
			color.YellowString("gsctl login <email> -e <endpoint>"))
	}

	// table headers
	output := []string{
		strings.Join([]string{
			color.CyanString("ENDPOINT URL"),
			color.CyanString("EMAIL"),
			color.CyanString("SELECTED"),
			color.CyanString("LOGGED IN"),
		}, "|"),
	}

	for endpoint := range config.Config.Endpoints {
		selected := "no"
		loggedIn := "no"
		email := "n/a"

		selectedEndpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
		if endpoint == selectedEndpoint {
			selected = "yes"
		}

		if config.Config.Endpoints[endpoint].Token != "" {
			loggedIn = "yes"
		}

		if config.Config.Endpoints[endpoint].Email != "" {
			email = config.Config.Endpoints[endpoint].Email
		}

		output = append(output, strings.Join([]string{endpoint, email, selected, loggedIn}, "|"))
	}

	return columnize.SimpleFormat(output)
}
