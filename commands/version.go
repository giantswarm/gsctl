package commands

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/util"
)

type updateAvailabilityInfo struct {
	currentVersion  string
	latestVersion   string
	updateAvailable bool
}

const (
	updateCheckTimeout = 2 * time.Second
)

var (
	// VersionCommand is the "version" go command
	VersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Print version number",
		Long: `Prints the gsctl version number.

When executed with the -v/--verbose flag, the build date is printed in addition.`,
		Run: printVersion,
	}
)

func init() {
	RootCommand.AddCommand(VersionCommand)
}

// printInfo prints some information on the current user and configuration
func printVersion(cmd *cobra.Command, args []string) {
	output := []string{}

	if config.Version != "" {
		output = append(output, color.YellowString("Version:")+"|"+color.CyanString(config.Version))
	} else {
		output = append(output, color.YellowString("Version:")+"|"+color.CyanString("n/a (version number is only available in a built binary)"))
	}
	if config.BuildDate != "" {
		output = append(output, color.YellowString("Build date:")+"|"+color.CyanString(config.BuildDate))
	} else {
		output = append(output, color.YellowString("Build date:")+"|"+color.CyanString("n/a (build date/time is only available in a built binary)"))
	}
	fmt.Println(columnize.SimpleFormat(output))

	// check for an update
	cv := currentVersion()
	if cv == "0.0.0" {
		return
	}

	info, err := checkUpdateAvailable(config.VersionCheckURL)
	if err == nil {
		// we are ignoring any errors from failed versionchecks
		// as we don't want to get into the way. And we only print this for
		// a properly built gsctl binary.
		config.Config.LastVersionCheck = time.Now()
		config.WriteToFile()
		if info.updateAvailable {
			fmt.Println()
			fmt.Println(formatUpdateInfo(info))
		}
	}
}

// latestVersion returns the latest available version as string
func latestVersion(url string) (string, error) {
	// an HTTP client that doesn't follow redirects and timeouts quickly
	// in order to not let the user wait too long
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(updateCheckTimeout),
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "", microerror.Mask(errors.UpdateCheckFailed)
	}

	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

// currentVersion returns the current gsctl version as string.
// When executed from a non-build (e. g. go test), it returns the
// equivalent of "0.0.0"
func currentVersion() string {
	if config.Version != "" {
		// remove '+git'
		v := strings.Replace(config.Version, "+git", "", 1)
		return v
	}
	return "0.0.0"
}

// checkUpdateAvailable checks whether an update is available and returns info as a struct
func checkUpdateAvailable(url string) (updateAvailabilityInfo, error) {
	current := currentVersion()
	latest, err := latestVersion(url)
	if err != nil {
		return updateAvailabilityInfo{}, microerror.Mask(err)
	}

	info := updateAvailabilityInfo{
		currentVersion:  current,
		latestVersion:   latest,
		updateAvailable: false,
	}

	comp, err := util.CompareVersions(latest, current)
	if err != nil {
		return updateAvailabilityInfo{}, microerror.Mask(err)
	}

	if comp > 0 {
		info.updateAvailable = true
	}

	return info, nil
}

// timeSinceLastVersionCheck returns the time sine the last update check
func timeSinceLastVersionCheck() time.Duration {
	return time.Since(config.Config.LastVersionCheck)
}

// versionCheckDue returns whether a version check is due
func versionCheckDue() bool {
	return timeSinceLastVersionCheck() > config.VersionCheckInterval
}

// formatUpdateInfo creates printable info about an available update
func formatUpdateInfo(info updateAvailabilityInfo) string {
	output := color.YellowString(fmt.Sprintf("Good news: an update for %s is available.\n", config.ProgramName))
	output += fmt.Sprintf("Please visit https://github.com/giantswarm/gsctl/releases/tag/%s for details.\n", info.latestVersion)
	return output
}
