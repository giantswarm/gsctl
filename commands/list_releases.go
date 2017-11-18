package commands

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

const (
	listReleasesActivityName = "list-releases"
)

var (

	// ListReleasesCommand performs the "list releases" function
	ListReleasesCommand = &cobra.Command{
		Use:   "releases",
		Short: "List releases to be used with clusters",
		Long: `Prints detail on all available releases.

A release is a software bundle that constitutes a cluster. It is identified by its semantic version number.`,
		PreRun: listReleasesValidationOutput,
		Run:    listReleasesOutput,
	}
)

// listReleasesArguments are the actual arguments used to call the
// listReleases() function.
type listReleasesArguments struct {
	apiEndpoint string
	token       string
}

// defaultListReleasesArguments returns a new listReleasesArguments struct
// based on global variables (= command line options from cobra).
func defaultListReleasesArguments() listReleasesArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	return listReleasesArguments{
		apiEndpoint: endpoint,
		token:       token,
	}
}

// listReleasesResult is the data structure returned by the listReleases() function.
type listReleasesResult struct {
	releases []gsclientgen.V4ReleaseListItem
}

func init() {
	ListCommand.AddCommand(ListReleasesCommand)
}

// listReleasesValidationOutput does our pre-checks and shows errors, in case
// something is missing.
func listReleasesValidationOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListReleasesArguments()
	err := listReleasesValidate(&args)
	if err != nil {
		var headline string
		var subtext string

		switch {
		case IsNotLoggedInError(err):
			headline = "You are not logged in."
			subtext = "Please log in using 'gsctl login <email>' or set an auth token as a command line argument."
			subtext += " See `gsctl list releases --help` for details."
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}
}

// listReleasesValidate validates our pre-conditions and returns an error in
// case something is missing.
func listReleasesValidate(args *listReleasesArguments) error {
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}

	return nil
}

// componentsString concatenates components and their version to a string.
func componentsString(components []gsclientgen.V4ReleaseComponent) string {
	items := []string{}

	slice.Sort(components[:], func(i, j int) bool {
		return components[i].Name < components[j].Name
	})

	for _, component := range components {
		items = append(items, component.Name+":"+component.Version)
	}

	return strings.Join(items, " ")
}

// listReleasesOutput is the function called to list releases and display
// errors in case they happen
func listReleasesOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListReleasesArguments()
	result, err := listReleases(args)

	// error output
	if err != nil {
		var headline string
		var subtext string

		switch {
		case IsNotLoggedInError(err):
			headline = "You are not logged in."
			subtext = "Please log in using 'gsctl login <email>' or set an auth token as a command line argument."
			subtext += " See `gsctl list releases --help` for details."
		case IsNotAuthorizedError(err):
			headline = "You are not authorized for this cluster."
			subtext = "You have no permission to access releases for this cluster. Please check your credentials."
		case IsInternalServerError(err):
			headline = "An internal error occurred."
			subtext = "Please notify the Giant Swarm support team, or try listing releases again in a few moments."
		case IsUnknownError(err):
			headline = "An error occurred."
			subtext = "Please notify the Giant Swarm support team, or try listing releases again in a few moments."
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
	if len(result.releases) == 0 {
		fmt.Println(color.RedString("No releases available."))
		fmt.Println("We cannot find any releases. Please contact the Giant Swarm support team to find out if there is a problem to be solved.")
		os.Exit(1)
	} else {

		for _, release := range result.releases {

			created := util.ParseDate(release.Timestamp)
			active := "false"
			if release.Active {
				active = "true"
			}

			// YAML-style output of all release details
			fmt.Println("---")
			fmt.Printf("%s %s\n", color.YellowString("Version:"), release.Version)
			fmt.Printf("%s %s\n", color.YellowString("Created:"), util.ShortDate(created))
			fmt.Printf("%s %s\n", color.YellowString("Active:"), active)
			fmt.Printf("%s\n", color.YellowString("Components:"))

			for _, component := range release.Components {
				fmt.Printf("  %s %s\n", color.YellowString(component.Name+":"), component.Version)
			}

			fmt.Printf("%s\n", color.YellowString("Changelog:"))

			for _, change := range release.Changelog {
				fmt.Printf("  %s %s\n", color.YellowString(change.Component+":"), change.Description)
			}

		}
	}
}

// listReleases fetches releases and returns them as a structured result.
func listReleases(args listReleasesArguments) (listReleasesResult, error) {
	result := listReleasesResult{}

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}

	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(couldNotCreateClientError)
	}
	authHeader := "giantswarm " + args.token
	releasesResponse, apiResponse, err := apiClient.GetReleases(authHeader,
		requestIDHeader, listReleasesActivityName, cmdLine)

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

	// sort releases by date
	if len(releasesResponse) > 1 {
		slice.Sort(releasesResponse[:], func(i, j int) bool {
			return releasesResponse[i].Timestamp > releasesResponse[j].Timestamp
		})
	}

	// sort changelog and components by component name
	for n := range releasesResponse {
		slice.Sort(releasesResponse[n].Components[:], func(i, j int) bool {
			return releasesResponse[n].Components[i].Name < releasesResponse[n].Components[j].Name
		})
		slice.Sort(releasesResponse[n].Changelog[:], func(i, j int) bool {
			return releasesResponse[n].Changelog[i].Component < releasesResponse[n].Changelog[j].Component
		})
	}

	result.releases = releasesResponse

	return result, nil
}
