package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	microerror "github.com/giantswarm/microkit/error"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

type updateAvailabilityInfo struct {
	currentVersion  *semver.Version
	latestVersion   *semver.Version
	updateAvailable bool
}

var (
	// VersionCommand is the "version" go command
	VersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Print version number",
		Long: `Prints the gsctl version number.

When executed with the --verbose flag, the build date is printed in addition.`,
		Run: printVersion,
	}
)

func init() {
	RootCommand.AddCommand(VersionCommand)
}

// printInfo prints some information on the current user and configuration
func printVersion(cmd *cobra.Command, args []string) {
	if config.Version != "" {
		fmt.Println(config.Version)
	} else {
		fmt.Println("Info: version number is only available in a built binary")
	}
	if cmdVerbose {
		if config.BuildDate != "" {
			fmt.Println(config.BuildDate)
		} else {
			fmt.Println("Info: build date/time is only available in a built binary")
		}
	}

	if versionCheckDue() {
		cv := currentVersion()
		if !cv.Equal(*semver.New("0.0.0")) {
			info, err := checkUpdateAvailable(config.VersionCheckURL)
			if err == nil {
				// we are ignoring any errors from failed versionchecks
				// as we don't want to get into the way. And we only print this for
				// a properly built gsctl binary.
				config.Config.LastVersionCheck = time.Now()
				config.WriteToFile()
				if info.updateAvailable {
					fmt.Println()
					fmt.Println(updateInfo(info))
				}
			}
		}
	}
}

// latestVersion returns the latest available version as semver.Version
func latestVersion(url string) (*semver.Version, error) {
	// to be unobstructive, we timeout quickly.
	timeout := time.Duration(2 * time.Second)
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return semver.New("0.0.0"), microerror.MaskAny(err)
	}
	defer resp.Body.Close()
	contentBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return semver.New("0.0.0"), microerror.MaskAny(err)
	}
	content := strings.TrimSpace(string(contentBytes))
	return semver.New(content), nil
}

// currentVersion returns the current gsctl version as semver.Version.
// When executed from a nun-build (e. g. go test), it returns the
// equivalent of "0.0.0"
func currentVersion() *semver.Version {
	if config.Version != "" {
		return semver.New(config.Version)
	}
	return semver.New("0.0.0")
}

// checkUpdateAvailable checks whether an update is available and returns info as a struct
func checkUpdateAvailable(url string) (updateAvailabilityInfo, error) {
	current := currentVersion()
	latest, err := latestVersion(url)
	if err != nil {
		return updateAvailabilityInfo{}, microerror.MaskAny(err)
	}

	info := updateAvailabilityInfo{
		currentVersion:  current,
		latestVersion:   latest,
		updateAvailable: false,
	}

	if current.LessThan(*latest) {
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

// updateInfo creates printable info about an available update
func updateInfo(info updateAvailabilityInfo) string {
	output := color.YellowString(fmt.Sprintf("Good news: an update for %s is available.\n", config.ProgramName))
	output += fmt.Sprintf("Please visit https://github.com/giantswarm/gsctl/releases/tag/%s for details.\n", info.latestVersion)
	return output
}
