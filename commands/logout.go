package commands

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
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
	return logoutArguments{
		apiEndpoint: config.Config.ChooseEndpoint(cmdAPIEndpoint),
		token:       cmdToken,
	}
}

func init() {
	RootCommand.AddCommand(LogoutCommand)
}

// TODO: separate validation and validation result output
func logoutValidationOutput(cmd *cobra.Command, args []string) {
	if config.Config.Token == "" && cmdToken == "" {
		fmt.Println("You weren't logged in here, but better be safe than sorry.")
		os.Exit(1)
	}
}

// logoutOutput performs our logout function and displays the result.
func logoutOutput(cmd *cobra.Command, extraArgs []string) {
	logoutArgs := defaultLogoutArguments()

	logoutArgs.token = config.Config.Token
	if cmdToken != "" {
		logoutArgs.token = cmdToken
	}

	err := logout(logoutArgs)
	if err != nil {
		var headline = ""
		var subtext = ""
		switch err.Error() {
		case "":
			return
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	fmt.Printf("You have logged out from endpoint %s.\n", color.CyanString(logoutArgs.apiEndpoint))
}

// logout terminates the current user session.
// The email and token are erased from the local config file.
// Returns nil in case of success, or an error otherwise.
func logout(args logoutArguments) error {
	// erase local credentials, no matter what the result on the API side is
	config.Config.Logout(config.Config.ChooseEndpoint(cmdAPIEndpoint))

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return microerror.Mask(couldNotCreateClientError)
	}

	authHeader := "giantswarm " + args.token
	logoutResponse, apiResponse, err := apiClient.UserLogout(authHeader, requestIDHeader, logoutActivityName, cmdLine)
	if err != nil {
		return fmt.Errorf("Error in API request to logout: %s", err.Error())
	}

	if logoutResponse.StatusCode != apischema.STATUS_CODE_RESOURCE_DELETED {
		if apiResponse.Response.StatusCode == http.StatusUnauthorized {
			// we ignore a 401 (Unauthorized) response here, as it means in most cases
			// that the token submitted was already expired.
			return nil
		}
		return fmt.Errorf("Error in API request to logout: %#v", logoutResponse)
	}

	return nil
}
