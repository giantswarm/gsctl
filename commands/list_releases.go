package commands

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client/clienterror"
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
	releases []*models.V4ReleaseListItem
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

			created := util.ParseDate(*release.Timestamp)
			active := "false"
			if release.Active {
				active = "true"
			}

			// YAML-style output of all release details
			fmt.Println("---")
			fmt.Printf("%s %s\n", color.YellowString("Version:"), *release.Version)
			fmt.Printf("%s %s\n", color.YellowString("Created:"), util.ShortDate(created))
			fmt.Printf("%s %s\n", color.YellowString("Active:"), active)
			fmt.Printf("%s\n", color.YellowString("Components:"))

			for _, component := range release.Components {
				fmt.Printf("  %s %s\n", color.YellowString(*component.Name+":"), *component.Version)
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

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = listReleasesActivityName

	response, err := ClientV2.GetReleases(auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode >= http.StatusInternalServerError {
				return result, microerror.Maskf(internalServerError, err.Error())
			} else if clientErr.HTTPStatusCode == http.StatusUnauthorized {
				return result, microerror.Mask(notAuthorizedError)
			}
		}

		return result, microerror.Mask(err)
	}

	// success

	// sort releases by version (descending)
	if len(response.Payload) > 1 {
		sort.Slice(response.Payload[:], func(i, j int) bool {
			vi := semver.New(*response.Payload[i].Version)
			vj := semver.New(*response.Payload[j].Version)
			return vi.LessThan(*vj)
		})
	}

	// sort changelog and components by component name
	for n := range response.Payload {
		sort.Slice(response.Payload[n].Components[:], func(i, j int) bool {
			if *response.Payload[n].Components[i].Name == "kubernetes" {
				return true
			}
			return *response.Payload[n].Components[i].Name < *response.Payload[n].Components[j].Name
		})
		sort.Slice(response.Payload[n].Changelog[:], func(i, j int) bool {
			if response.Payload[n].Changelog[i].Component == "kubernetes" {
				return true
			}
			return response.Payload[n].Changelog[i].Component < response.Payload[n].Changelog[j].Component
		})
	}

	result.releases = response.Payload

	return result, nil
}
