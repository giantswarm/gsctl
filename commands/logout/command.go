// Package logout implements the logout command.
package logout

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/microerror"
)

const (
	logoutActivityName = "logout"
)

var (
	// Command performs a logout
	Command = &cobra.Command{
		Use:   "logout",
		Short: "Sign the current user out",
		Long: `Terminates the user's session with the current endpoint and invalidates the authentication token.

If an endpoint was selected before, it remains selected. Re-login using 'gsctl login <email>'.`,
		PreRun: printValidation,
		Run:    printResult,
	}
)

// Arguments holds all arguments that can influence our business function.
type Arguments struct {
	// apiEndpoint is the API to log out from
	apiEndpoint string
	// token is the session token to expire (log out)
	token             string
	userProvidedToken string
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		token:             token,
		userProvidedToken: flags.Token,
	}
}

func printValidation(cmd *cobra.Command, args []string) {
	if config.Config.Token == "" && flags.Token == "" {
		fmt.Println("You weren't logged in here, but better be safe than sorry.")
		os.Exit(1)
	}
}

// printResult performs our logout function and displays the result.
func printResult(cmd *cobra.Command, extraArgs []string) {
	logoutArgs := collectArguments()

	err := logout(logoutArgs)

	if err != nil {

		// Special treatment: We ignore the fact that the user was not logged in
		// and act as if she just logged out.
		if clienterror.IsUnauthorizedError(err) {
			fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
			os.Exit(0)
		}

		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		// handle non-common errors
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
}

// logout terminates the current user session.
// The email and token are erased from the local config file.
func logout(args Arguments) error {
	// erase local credentials, no matter what the result on the API side is
	defer config.Config.Logout(args.apiEndpoint)

	if config.Config.Scheme == "Bearer" {
		return nil
	}

	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return microerror.Mask(err)
	}

	ap := clientWrapper.DefaultAuxiliaryParams()
	ap.ActivityName = logoutActivityName

	_, err = clientWrapper.DeleteAuthToken(args.token, ap)
	return microerror.Mask(err)
}
