package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsctl/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// listEndpointsArgs are the arguments we pass to the actual functions
// listing endpoints and printing endpoints lists
type listEndpointsArguments struct {
	apiEndpoint string
	token       string
}

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

// defaultListEndpointArgs returns listEndpointsArguments
// with settings laoded from flags etc.
func defaultListEndpointArguments() listEndpointsArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	return listEndpointsArguments{
		apiEndpoint: endpoint,
		token:       token,
	}
}

// listEndpoints prints a table with all endpoint URLs the user has used
func listEndpoints(cmd *cobra.Command, args []string) {
	myArgs := defaultListEndpointArguments()
	output := endpointsTable(myArgs)
	if output != "" {
		fmt.Println(output)
	}
}

// endpointsTable returns a table of clusters the user has access to
func endpointsTable(args listEndpointsArguments) string {
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

	// get keys (URLs) and sort by them
	endpointURLs := make([]string, 0, len(config.Config.Endpoints))
	for u := range config.Config.Endpoints {
		endpointURLs = append(endpointURLs, u)
	}

	sort.Slice(endpointURLs, func(i, j int) bool {
		return endpointURLs[i] < endpointURLs[j]
	})

	for _, endpoint := range endpointURLs {
		selected := "no"
		loggedIn := "no"
		email := "n/a"

		if endpoint == args.apiEndpoint {
			selected = "yes"
		}

		if config.Config.Endpoints[endpoint].Token != "" {
			loggedIn = "yes"
		}

		if config.Config.Endpoints[endpoint].Email != "" {
			email = config.Config.Endpoints[endpoint].Email
		}

		row := ""
		if endpoint == args.apiEndpoint {
			// highlight if selected
			row = strings.Join([]string{
				color.YellowString(endpoint),
				color.YellowString(email),
				color.YellowString(selected),
				color.YellowString(loggedIn),
			}, "|")
		} else {
			row = strings.Join([]string{endpoint, email, selected, loggedIn}, "|")
		}
		output = append(output, row)
	}

	return columnize.SimpleFormat(output)
}
