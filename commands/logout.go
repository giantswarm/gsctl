package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

const (
	logoutActivityName = "login"
)

var (
	// LogoutCommand performs a logout
	LogoutCommand = &cobra.Command{
		Use:   "logout",
		Short: "Sign the current user out",
		Long: `Terminates the user's session with the current endpoint and invalidates the authentication token.

If an endpoint was selected before, it remains selected. Re-login using 'gsctl login <email>'.`,
		PreRun: logoutValidationOutput,
		Run:    logoutOutput,
	}
)

type logoutArguments struct {
	// apiEndpoint is the API to log out from
	apiEndpoint string
	// token is the session token to expire (log out)
	token string
}

func defaultLogoutArguments() logoutArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	return logoutArguments{
		apiEndpoint: endpoint,
		token:       token,
	}
}

func init() {
	RootCommand.AddCommand(LogoutCommand)
}

func logoutValidationOutput(cmd *cobra.Command, args []string) {
	if config.Config.Token == "" && cmdToken == "" {
		fmt.Println("You weren't logged in here, but better be safe than sorry.")
		os.Exit(1)
	}
}

// logoutOutput performs our logout function and displays the result.
func logoutOutput(cmd *cobra.Command, extraArgs []string) {
	logoutArgs := defaultLogoutArguments()

	err := logout(logoutArgs)

	if err != nil {

		// Special treatment: We ignore the fact that the user was not logged in
		// and act as if she just logged out.
		if IsNotAuthorizedError(err) {
			fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
			os.Exit(0)
		}

		handleCommonErrors(err)

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

	_, err := ClientV2.DeleteAuthToken(args.token)
	return err
}
