// Package logout implements the logout command.
package logout

import (
	"fmt"
	"net/http"
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

type logoutArguments struct {
	// apiEndpoint is the API to log out from
	apiEndpoint string
	// token is the session token to expire (log out)
	token string
}

func defaultLogoutArguments() logoutArguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)

	return logoutArguments{
		apiEndpoint: endpoint,
		token:       token,
	}
}

func printValidation(cmd *cobra.Command, args []string) {
	if config.Config.Token == "" && flags.CmdToken == "" {
		fmt.Println("You weren't logged in here, but better be safe than sorry.")
		os.Exit(1)
	}
}

// printResult performs our logout function and displays the result.
func printResult(cmd *cobra.Command, extraArgs []string) {
	logoutArgs := defaultLogoutArguments()

	err := logout(logoutArgs)

	if err != nil {

		// Special treatment: We ignore the fact that the user was not logged in
		// and act as if she just logged out.
		if clientError, ok := err.(*clienterror.APIError); ok {
			if clientError.HTTPStatusCode == http.StatusUnauthorized {
				fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
				os.Exit(0)
			}
		}

		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		// handle non-common errors
		fmt.Println(color.RedString(err.Error()))
		os.Exit(1)
	}

	fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
}

// logout terminates the current user session.
// The email and token are erased from the local config file.
func logout(args logoutArguments) error {
	// erase local credentials, no matter what the result on the API side is
	defer config.Config.Logout(args.apiEndpoint)

	if config.Config.Scheme == "Bearer" {
		return nil
	}

	clientV2, err := client.NewWithConfig(flags.CmdAPIEndpoint, flags.CmdToken)
	if err != nil {
		return microerror.Mask(err)
	}

	ap := clientV2.DefaultAuxiliaryParams()
	ap.ActivityName = logoutActivityName

	_, err = clientV2.DeleteAuthToken(args.token, ap)
	return microerror.Mask(err)
}
