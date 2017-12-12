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
// with settings loaded from flags etc.
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

	// get keys (URLs) and sort by them
	endpointURLs := make([]string, 0, len(config.Config.Endpoints))
	for u := range config.Config.Endpoints {
		endpointURLs = append(endpointURLs, u)
	}

	// detect if we want to show the alias column
	hasAlias := false
	for _, endpoint := range endpointURLs {
		if config.Config.Endpoints[endpoint].Alias != "" {
			hasAlias = true
		}
	}

	// sort by alias first, endpoint URL second
	sort.Slice(endpointURLs, func(i, j int) bool {
		return endpointURLs[i] < endpointURLs[j]
	})
	sort.Slice(endpointURLs, func(i, j int) bool {
		aliasi := config.Config.Endpoints[endpointURLs[i]].Alias
		aliasj := config.Config.Endpoints[endpointURLs[j]].Alias
		// sort empty alias to bottom position
		if aliasi == "" {
			aliasi = "zzzzz"
		}
		if aliasj == "" {
			aliasj = "zzzzz"
		}
		return aliasi < aliasj
	})

	// table headers
	output := []string{}
	headers := []string{}

	if hasAlias {
		headers = append(headers, color.CyanString("ALIAS"))
	}
	headers = append(headers, color.CyanString("ENDPOINT URL"))
	headers = append(headers, color.CyanString("EMAIL"))
	headers = append(headers, color.CyanString("SELECTED"))
	headers = append(headers, color.CyanString("LOGGED IN"))
	output = append(output, strings.Join(headers, "|"))

	for _, endpoint := range endpointURLs {
		selected := "no"
		loggedIn := "no"
		email := "n/a"
		alias := "n/a"

		if config.Config.Endpoints[endpoint].Alias != "" {
			alias = config.Config.Endpoints[endpoint].Alias
		}

		if endpoint == args.apiEndpoint {
			selected = "yes"
		}

		if config.Config.Endpoints[endpoint].Token != "" {
			loggedIn = "yes"
		}

		if config.Config.Endpoints[endpoint].Email != "" {
			email = config.Config.Endpoints[endpoint].Email
		}

		columns := []string{}
		if endpoint == args.apiEndpoint {
			// highlight if selected
			if hasAlias {
				columns = append(columns, color.YellowString(alias))
			}
			columns = append(columns, color.YellowString(endpoint))
			columns = append(columns, color.YellowString(email))
			columns = append(columns, color.YellowString(selected))
			columns = append(columns, color.YellowString(loggedIn))
		} else {
			if hasAlias {
				columns = append(columns, alias)
			}
			columns = append(columns, endpoint)
			columns = append(columns, email)
			columns = append(columns, selected)
			columns = append(columns, loggedIn)
		}
		output = append(output, strings.Join(columns, "|"))
	}

	return columnize.SimpleFormat(output)
}
