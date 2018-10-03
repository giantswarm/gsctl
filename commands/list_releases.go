package commands

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
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
	releases, err := listReleases(args)

	if err != nil {
		handleCommonErrors(err)

		if clientErr, ok := err.(*clienterror.APIError); ok {
			fmt.Println(color.RedString(clientErr.ErrorMessage))
			if clientErr.ErrorDetails != "" {
				fmt.Println(clientErr.ErrorDetails)
			}
		} else {
			fmt.Println(color.RedString("Error: %s", err.Error()))
		}
		os.Exit(1)
	}

	// success
	if len(releases) == 0 {
		fmt.Println(color.RedString("No releases available."))
		fmt.Println("We cannot find any releases. Please contact the Giant Swarm support team to find out if there is a problem to be solved.")
	}

	// table headers
	output := []string{strings.Join([]string{
		color.CyanString("VERSION"),
		color.CyanString("STATUS"),
		color.CyanString("KUBERNETES"),
		color.CyanString("CONTAINERLINUX"),
		color.CyanString("COREDNS"),
		color.CyanString("CALICO"),
		color.CyanString("CREATED"),
	}, "|")}

	var major int64
	var status string
	major = 0
	status = "-"

	for _, release := range releases {
		created := util.ShortDate(util.ParseDate(*release.Timestamp))
		kubernetes_version := "n/a"
		container_linux_version := "n/a"
		coredns_version := "n/a"
		calico_version := "n/a"

		// as long as the status information is not specific in the API
		// we start with deprecated, find the active one and then switch
		// to "wip" for each major version
		version, err := semver.NewVersion(*release.Version)

		if err == nil {
			if version.Major() > major {
				major = version.Major()
				status = "deprecated"
			}

			if release.Active {
				status = "active"
			} else if status == "active" {
				status = "wip"
			}
		} else {
			// release version couldn't be parsed
			major = 0
			status = "-"
		}

		for _, component := range release.Components {
			if *component.Name == "kubernetes" {
				kubernetes_version = *component.Version
			}
			if *component.Name == "containerlinux" {
				container_linux_version = *component.Version
			}
			if *component.Name == "coredns" {
				coredns_version = *component.Version
			}
			if *component.Name == "calico" {
				calico_version = *component.Version
			}
		}

		if status == "active" {
			output = append(output, strings.Join([]string{
				color.YellowString(*release.Version),
				color.YellowString(status),
				color.YellowString(kubernetes_version),
				color.YellowString(container_linux_version),
				color.YellowString(coredns_version),
				color.YellowString(calico_version),
				color.YellowString(created),
			}, "|"))
		} else {
			output = append(output, strings.Join([]string{
				*release.Version,
				status,
				kubernetes_version,
				container_linux_version,
				coredns_version,
				calico_version,
				created,
			}, "|"))
		}
	}

	fmt.Println(columnize.SimpleFormat(output))
}

// listReleases fetches releases and returns them as a structured result.
func listReleases(args listReleasesArguments) ([]*models.V4ReleaseListItem, error) {
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = listReleasesActivityName

	response, err := ClientV2.GetReleases(auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode >= http.StatusInternalServerError {
				return nil, microerror.Maskf(internalServerError, err.Error())
			} else if clientErr.HTTPStatusCode == http.StatusUnauthorized {
				return nil, microerror.Mask(notAuthorizedError)
			}
		}

		return nil, microerror.Mask(err)
	}

	// sort releases by version (ascending)
	sort.Slice(response.Payload[:], func(i, j int) bool {
		vi, err := semver.NewVersion(*response.Payload[i].Version)
		if err != nil {
			return false
		}
		vj, err := semver.NewVersion(*response.Payload[j].Version)
		if err != nil {
			return true
		}

		return vj.GreaterThan(vi)
	})

	return response.Payload, nil
}
