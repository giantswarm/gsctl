package commands

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
)

const (
	logoutActivityName = "login"

	// errors
	errInvalidToken = "submitted token was not valid"
)

var (
	// LogoutCommand performs a logout
	LogoutCommand = &cobra.Command{
		Use:     "logout",
		Short:   "Sign the current user out",
		Long:    `This will terminate the current user's session and invalidate the authentication token.`,
		PreRunE: logoutValidationOutput,
		Run:     logoutOutput,
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
		apiEndpoint: cmdAPIEndpoint,
		token:       cmdToken,
	}
}

func init() {
	RootCommand.AddCommand(LogoutCommand)
}

// TODO: separate validation and validation result output
func logoutValidationOutput(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in")
	}
	return nil
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
}

// logout terminates the current user session.
// The email and token are erased from the local config file.
// Returns nil in case of success, or an error otherwise.
func logout(args logoutArguments) error {
	// erase local credentials, no matter what the result on the API side is
	config.Config.Token = ""
	config.Config.Email = ""
	config.WriteToFile()

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)

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
