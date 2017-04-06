package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

const (
	logoutActivityName string = "login"
)

var (
	// LogoutCommand performs a logout
	LogoutCommand = &cobra.Command{
		Use:     "logout",
		Short:   "Sign the current user out",
		Long:    `This will terminate the current user's session and invalidate the authentication token.`,
		PreRunE: checkLogout,
		Run:     logout,
	}
)

func init() {
	RootCommand.AddCommand(LogoutCommand)
}

func checkLogout(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in")
	}
	return nil
}

func logout(cmd *cobra.Command, args []string) {

	currentToken := config.Config.Token
	if cmdToken != "" {
		currentToken = cmdToken
	}

	// erase local credentials in any case
	config.Config.Token = ""
	config.Config.Email = ""
	config.WriteToFile()

	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)

	authHeader := "giantswarm " + currentToken
	logoutResponse, apiResponse, err := client.UserLogout(authHeader, requestIDHeader, logoutActivityName, cmdLine)
	if err != nil {
		fmt.Println("Info: The client doesn't handle the API's 401 response yet.")
		fmt.Println("Seeing this error likely means: The passed token was no longer valid.")
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}
	if logoutResponse.StatusCode == apischema.STATUS_CODE_RESOURCE_DELETED {
		// remove token from settings
		// unless we unathenticated the token from flags
		fmt.Println(color.GreenString("Successfully logged out"))
	} else {
		fmt.Println(color.RedString("Unhandled response code: %s", logoutResponse.StatusCode))
		dumpAPIResponse(*apiResponse)
	}
}
