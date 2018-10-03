package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	ShowReleaseCommand = &cobra.Command{
		Use:   "release",
		Short: "Show release details",
		Long: `Display details of a release

Examples:

  gsctl show release 4.2.2
`,

		// PreRun checks a few general things, like authentication.
		PreRun: showReleasePreRunOutput,

		// Run calls the business function and prints results and errors.
		Run: showReleaseRunOutput,
	}
)

const (
	showReleaseActivityName = "show-release"
)

type showReleaseArguments struct {
	apiEndpoint    string
	authToken      string
	scheme         string
	releaseVersion string
	verbose        bool
}

func defaultShowReleaseArguments() showReleaseArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return showReleaseArguments{
		apiEndpoint:    endpoint,
		authToken:      token,
		scheme:         scheme,
		releaseVersion: "",
		verbose:        cmdVerbose,
	}
}

func init() {
	ShowCommand.AddCommand(ShowReleaseCommand)
}

func showReleasePreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowReleaseArguments()
	err := verifyShowReleasePreconditions(args, cmdLineArgs)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	// handle non-common errors
	fmt.Println(color.RedString(err.Error()))
	os.Exit(1)
}

func verifyShowReleasePreconditions(args showReleaseArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	if len(cmdLineArgs) == 0 {
		return microerror.Mask(releaseVersionMissingError)
	}
	return nil
}

// getReleaseDetails fetches release details from the API
func getReleaseDetails(releaseVersion, scheme, token, endpoint string) (*models.V4ReleaseListItem, error) {

	// perform API call
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

	for _, release := range response.Payload {
		if *release.Version == releaseVersion {
			return release, nil
		}
	}

	return nil, microerror.Mask(releaseNotFoundError)
}

// showReleaseRunOutput prints the release information on stdout
func showReleaseRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultShowReleaseArguments()
	args.releaseVersion = cmdLineArgs[0]
	release, err := getReleaseDetails(args.releaseVersion, args.scheme,
		args.authToken, args.apiEndpoint)

	// error output
	if err != nil {
		var headline = ""
		var subtext = ""

		// TODO: handle common and specific errors
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
