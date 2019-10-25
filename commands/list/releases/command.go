// Package releases implements the 'list releases' sub-command.
package releases

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/util"
)

const (
	listReleasesActivityName = "list-releases"

	outputFormatJSON  = "json"
	outputFormatTable = "table"

	outputJSONPrefix = ""
	outputJSONIndent = "  "
)

var (
	// Command performs the "list releases" function
	Command = &cobra.Command{
		Use:   "releases",
		Short: "List releases to be used with clusters",
		Long: `Prints detail on all available releases.

A release is a software bundle that constitutes a cluster. It is identified by its semantic version number.`,
		PreRun: printValidation,
		Run:    printResult,
	}

	cmdOutput string
)

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().StringVarP(&cmdOutput, "output", "o", "table", "Use 'json' for JSON output.")
}

// Arguments are the actual arguments used to call the
// listReleases() function.
type Arguments struct {
	apiEndpoint       string
	outputFormat      string
	scheme            string
	token             string
	userProvidedToken string
}

// collectArguments returns a new Arguments struct
// based on global variables (= command line options from cobra).
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		outputFormat:      cmdOutput,
		token:             token,
		scheme:            scheme,
		userProvidedToken: flags.Token,
	}
}

// printValidation does our pre-checks and shows errors, in case
// something is missing.
func printValidation(cmd *cobra.Command, extraArgs []string) {
	args := collectArguments()
	err := listReleasesPreconditions(&args)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)

	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

// listReleasesPreconditions validates our pre-conditions and returns an error in
// case something is missing.
func listReleasesPreconditions(args *Arguments) error {
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.outputFormat != outputFormatJSON && args.outputFormat != outputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, fmt.Sprintf("Output format '%s' is unknown", args.outputFormat))
	}

	return nil
}

// printResult is the function called to list releases and display
// errors in case they happen
func printResult(cmd *cobra.Command, extraArgs []string) {
	args := collectArguments()
	releases, err := listReleases(args)

	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

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

	if args.outputFormat == "json" {
		outputBytes, err := json.MarshalIndent(releases, outputJSONPrefix, outputJSONIndent)
		if err != nil {
			fmt.Println(color.RedString("Error while encoding JSON"))
			fmt.Printf("Details: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println(string(outputBytes))
		return
	}

	// success
	if len(releases) == 0 {
		fmt.Println(color.RedString("No releases available."))
		fmt.Println("We cannot find any releases. Please contact the Giant Swarm support team to find out if there is a problem to be solved.")
		return
	}

	// table headers
	output := []string{strings.Join([]string{
		color.CyanString("VERSION"),
		color.CyanString("STATUS"),
		color.CyanString("CREATED"),
		color.CyanString("KUBERNETES"),
		color.CyanString("CONTAINERLINUX"),
		color.CyanString("COREDNS"),
		color.CyanString("CALICO"),
	}, "|")}

	var major int64
	var status string
	major = 0
	status = "deprecated"

	for i, release := range releases {
		created := util.ShortDate(util.ParseDate(*release.Timestamp))
		kubernetesVersion := "n/a"
		containerLinuxVersion := "n/a"
		coreDNSVersion := "n/a"
		calicoVersion := "n/a"

		// As long as the status information is not specific in the API
		// we start with deprecated, find the active one and then switch
		// to "wip" for each major version.
		version, err := semver.NewVersion(*release.Version)

		if err == nil {
			if version.Major() > major {
				// Found new major release.
				major = version.Major()

				// If this is a new major version and the last release
				// likelihood is high that this is a wip release and
				// not deprecated.
				if i == len(releases)-1 {
					status = "wip"
				} else {
					status = "deprecated"
				}
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
				kubernetesVersion = *component.Version
			}
			if *component.Name == "containerlinux" {
				containerLinuxVersion = *component.Version
			}
			if *component.Name == "coredns" {
				coreDNSVersion = *component.Version
			}
			if *component.Name == "calico" {
				calicoVersion = *component.Version
			}
		}

		if status == "active" {
			output = append(output, strings.Join([]string{
				color.YellowString(*release.Version),
				color.YellowString(status),
				color.YellowString(created),
				color.YellowString(kubernetesVersion),
				color.YellowString(containerLinuxVersion),
				color.YellowString(coreDNSVersion),
				color.YellowString(calicoVersion),
			}, "|"))
		} else {
			output = append(output, strings.Join([]string{
				*release.Version,
				status,
				created,
				kubernetesVersion,
				containerLinuxVersion,
				coreDNSVersion,
				calicoVersion,
			}, "|"))
		}

		fmt.Println(columnize.SimpleFormat(output))
	}
}

// listReleases fetches releases and returns them as a structured result.
func listReleases(args Arguments) ([]*models.V4ReleaseListItem, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)

	if err != nil {
		return nil, microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listReleasesActivityName

	response, err := clientWrapper.GetReleases(auxParams)
	if err != nil {
		// create specific error types for cases we care about
		if clienterror.IsInternalServerError(err) {
			return nil, microerror.Maskf(errors.InternalServerError, err.Error())
		}
		if clienterror.IsUnauthorizedError(err) {
			return nil, microerror.Mask(errors.NotAuthorizedError)
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
