// Package release implements the 'show release' command.
package release

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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

var (
	// ShowReleaseCommand is the cobra command for 'gsctl show release'
	ShowReleaseCommand = &cobra.Command{
		Use:   "release",
		Short: "Show release details",
		Long: `Display details of a release

Examples:

  gsctl show release 4.2.2
`,

		// PreRun checks a few general things, like authentication.
		PreRun: printValidation,

		// Run calls the business function and prints results and errors.
		Run: printResult,
	}

	arguments Arguments
)

const (
	showReleaseActivityName = "show-release"
)

type Arguments struct {
	apiEndpoint       string
	authToken         string
	releaseVersion    string
	scheme            string
	userProvidedToken string
	verbose           bool
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		scheme:            scheme,
		releaseVersion:    "",
		userProvidedToken: flags.Token,
		verbose:           flags.Verbose,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	arguments = collectArguments()
	err := verifyShowReleasePreconditions(arguments, cmdLineArgs)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyShowReleasePreconditions(args Arguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(errors.ReleaseVersionMissingError)
	}
	return nil
}

// getReleaseDetails fetches release details from the API
func getReleaseDetails(args Arguments) (*models.V4ReleaseListItem, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// perform API call
	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = showReleaseActivityName

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

	for _, release := range response.Payload {
		if *release.Version == args.releaseVersion {
			return release, nil
		}
	}

	return nil, microerror.Mask(errors.ReleaseNotFoundError)
}

// printResult prints the release information on stdout
func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	arguments.releaseVersion = cmdLineArgs[0]
	release, err := getReleaseDetails(arguments)

	// error output
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		// TODO: handle specific errors
		switch {
		case err.Error() == "":
			return
		default:
			headline = err.Error()
		}

		// Print error output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	// success output
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
