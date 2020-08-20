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
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/util"
)

const listReleasesActivityName = "list-releases"

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

	arguments Arguments
)

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()

	Command.Flags().StringVarP(&flags.OutputFormat, "output", "o", formatting.OutputFormatTable, fmt.Sprintf("Use '%s' for JSON output. Defaults to human-friendly table output.", formatting.OutputFormatJSON))
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
		outputFormat:      flags.OutputFormat,
		token:             token,
		scheme:            scheme,
		userProvidedToken: flags.Token,
	}
}

// printValidation does our pre-checks and shows errors, in case
// something is missing.
func printValidation(cmd *cobra.Command, extraArgs []string) {
	arguments = collectArguments()
	err := listReleasesPreconditions(&arguments)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

// listReleasesPreconditions validates our pre-conditions and returns an error in
// case something is missing.
func listReleasesPreconditions(args *Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.token == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.outputFormat != formatting.OutputFormatJSON && args.outputFormat != formatting.OutputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, fmt.Sprintf("Output format '%s' is unknown", args.outputFormat))
	}

	return nil
}

// printResult is the function called to list releases and display
// errors in case they happen
func printResult(cmd *cobra.Command, extraArgs []string) {
	releases, err := listReleases(arguments)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

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

	if arguments.outputFormat == formatting.OutputFormatJSON {
		outputBytes, err := json.MarshalIndent(releases, formatting.OutputJSONPrefix, formatting.OutputJSONIndent)
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

	for _, release := range releases {
		created := util.ShortDate(util.ParseDate(*release.Timestamp))
		kubernetesVersion := "n/a"
		containerLinuxVersion := "n/a"
		coreDNSVersion := "n/a"
		calicoVersion := "n/a"

		status := "inactive"
		if release.Active {
			status = "active"
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
	}

	fmt.Println(columnize.SimpleFormat(output))
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
