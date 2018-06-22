package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bradfitz/slice"
	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

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
		PreRun: listReleasesPreRunOutput,
		Run:    listReleasesRunOutput,
	}
)

// listReleasesArguments are the actual arguments used to call the
// listReleases() function.
type listReleasesArguments struct {
	apiEndpoint string
	token       string
	scheme      string
}

// defaultListReleasesArguments returns a new listReleasesArguments struct
// based on global variables (= command line options from cobra).
func defaultListReleasesArguments() listReleasesArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return listReleasesArguments{
		apiEndpoint: endpoint,
		token:       token,
		scheme:      scheme,
	}
}

// listReleasesResult is the data structure returned by the listReleases() function.
type listReleasesResult struct {
	releases []gsclientgen.V4ReleaseListItem
}

func init() {
	ListCommand.AddCommand(ListReleasesCommand)
}

// listReleasesPreRunOutput does our pre-checks and shows errors, in case
// something is missing.
func listReleasesPreRunOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListReleasesArguments()
	err := listReleasesPreconditions(&args)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

// listReleasesPreconditions validates our pre-conditions and returns an error in
// case something is missing.
func listReleasesPreconditions(args *listReleasesArguments) error {
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(notLoggedInError)
	}

	return nil
}

// listReleasesRunOutput is the function called to list releases and display
// errors in case they happen
func listReleasesRunOutput(cmd *cobra.Command, extraArgs []string) {
	args := defaultListReleasesArguments()
	result, err := listReleases(args)

	// error output
	if err != nil {
		handleCommonErrors(err)

		var headline = err.Error()

		fmt.Println(color.RedString(headline))
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

	releasesResponse, apiResponse, err := Client.GetReleases(ClientConfig.AuthHeader,
		requestIDHeader, listReleasesActivityName, cmdLine)
	if err != nil {
		if apiResponse == nil || apiResponse.Response == nil {
			return result, microerror.Mask(noResponseError)
		}
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

	// sort releases by version (descending)
	if len(releasesResponse) > 1 {
		slice.Sort(releasesResponse[:], func(i, j int) bool {
			vi := semver.New(releasesResponse[i].Version)
			vj := semver.New(releasesResponse[j].Version)
			return vj.LessThan(*vi)
		})
	}

	// sort changelog and components by component name
	for n := range releasesResponse {
		slice.Sort(releasesResponse[n].Components[:], func(i, j int) bool {
			if releasesResponse[n].Components[i].Name == "kubernetes" {
				return true
			}
			return releasesResponse[n].Components[i].Name < releasesResponse[n].Components[j].Name
		})
		slice.Sort(releasesResponse[n].Changelog[:], func(i, j int) bool {
			if releasesResponse[n].Changelog[i].Component == "kubernetes" {
				return true
			}
			return releasesResponse[n].Changelog[i].Component < releasesResponse[n].Changelog[j].Component
		})
	}

	result.releases = releasesResponse

	return result, nil
}
