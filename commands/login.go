package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
)

const (
	loginActivityName = "login"
)

var (
	// cmdPassword is the password given via command line flag
	cmdPassword string

	// email address passed as a positional argument
	cmdEmail string

	// LoginCommand is the "login" CLI command
	LoginCommand = &cobra.Command{
		Use:   "login <email> [-e|--endpoint <endpoint>]",
		Short: "Sign in as a user",
		Long: `Sign in against an endpoint with email address and password.

This will select the given endpoint for subsequent commands.

The password has to be entered interactively or given as -p / --password flag.

The -e or --endpoint argument can be omitted if an endpoint is already selected.`,
		Example: "  gsctl login user@example.com --endpoint api.example.com",
		PreRun:  loginPreRunOutput,
		Run:     loginRunOutput,
	}
)

type loginResult struct {
	// apiEndpoint is the API endpoint the user has been logged in to
	apiEndpoint string
	// alias is the alternative, user friendly name for an endpoint
	alias string
	// loggedOutBefore is true if the user has been logged out from a previous session
	loggedOutBefore bool
	// endpointSwitched is true when the endpoint has been changed during login
	endpointSwitched bool
	// email is the email address we are signed in with
	email string
	// token is the new session token received
	token string
}

type loginArguments struct {
	apiEndpoint string
	email       string
	password    string
	verbose     bool
}

func defaultLoginArguments() loginArguments {
	return loginArguments{
		apiEndpoint: config.Config.ChooseEndpoint(cmdAPIEndpoint),
		email:       cmdEmail,
		password:    cmdPassword,
		verbose:     cmdVerbose,
	}
}

func init() {
	LoginCommand.Flags().StringVarP(&cmdPassword, "password", "p", "", "Password. If not given, will be prompted interactively.")
	RootCommand.AddCommand(LoginCommand)
}

// loginPreRunOutput runs our pre-checks.
// If an error occurred, it prints the error info and exits with non-zero code.
func loginPreRunOutput(cmd *cobra.Command, positionalArgs []string) {
	err := verifyLoginPreconditions(positionalArgs)

	if err != nil {
		var headline = ""
		var subtext = ""
		switch {
		case err.Error() == "":
			return
		case IsNoEmailArgumentGivenError(err):
			headline = "The email argument is required."
			subtext = "Please execute the command as 'gsctl login <email>'. See 'gsctl login --help' for details."
		case IsTokenArgumentNotApplicableError(err):
			headline = "The '--auth-token' flag cannot be used with the 'gsctl login' command."
		case IsEmptyPasswordError(err):
			headline = "The password cannot be empty."
			subtext = "Please call the command again and enter a non-empty password. See 'gsctl login --help' for details."
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

// verifyLoginPreconditions does the pre-checks and returns an error in case something's wrong.
func verifyLoginPreconditions(positionalArgs []string) error {
	if len(positionalArgs) >= 1 {
		// set cmdEmail for later use, as cobra doesn't do that for us
		cmdEmail = positionalArgs[0]
	} else {
		return microerror.Mask(noEmailArgumentGivenError)
	}

	// using auth token flag? The 'login' command is the only exception
	// where we can't accept this argument.
	if cmdToken != "" {
		return microerror.Mask(tokenArgumentNotApplicableError)
	}

	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)

	// interactive password prompt
	if cmdPassword == "" {
		fmt.Printf("Password for %s on %s: ", color.CyanString(cmdEmail), color.CyanString(endpoint))
		password, err := gopass.GetPasswd()
		if err != nil {
			return err
		}
		if string(password) == "" {
			return microerror.Mask(emptyPasswordError)
		}
		cmdPassword = string(password)
	}

	return nil
}

// loginRunOutput executes the login logic and
// prints output and sets the exit code.
func loginRunOutput(cmd *cobra.Command, args []string) {
	loginArgs := defaultLoginArguments()

	result, err := login(loginArgs)
	if err != nil {
		var headline = ""
		var subtext = ""
		switch {
		case err.Error() == "":
			return
		case client.IsEndpointNotSpecifiedError(err):
			headline = "No endpoint has been specified."
			subtext = "Please use the '-e|--endpoint' flag."
		case IsEmptyPasswordError(err):
			headline = "Empty password submitted"
			subtext = "The API server complains about the password provided."
			subtext += " Please make sure to provide a string with more than white space characters."
		case IsInvalidCredentialsError(err):
			headline = "Bad password or email address."
			subtext = fmt.Sprintf("Could not log you in to %s.", color.CyanString(loginArgs.apiEndpoint))
			subtext += " The email or the password provided (or both) was incorrect."
		default:
			headline = err.Error()
		}

		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	if result.loggedOutBefore && loginArgs.verbose {
		fmt.Println("You have been logged out from your previous session.")
	}

	if result.endpointSwitched {
		fmt.Printf("Endpoint selected: %s\n", result.apiEndpoint)
	}

	fmt.Println(color.GreenString("You are logged in as %s at %s.",
		result.email, result.apiEndpoint))
}

// login executes the authentication logic.
// If the user was logged in before, a logout is performed first.
func login(args loginArguments) (loginResult, error) {
	result := loginResult{
		apiEndpoint:      args.apiEndpoint,
		email:            args.email,
		loggedOutBefore:  false,
		endpointSwitched: false,
	}

	endpointBefore := config.Config.SelectedEndpoint
	if result.apiEndpoint != endpointBefore {
		result.endpointSwitched = true
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(args.password))

	// log out if logged in
	if config.Config.Token != "" {
		result.loggedOutBefore = true
		// we deliberately ignore the logout result here
		logout(logoutArguments{
			apiEndpoint: args.apiEndpoint,
			token:       config.Config.Token,
		})
	}

	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(clientErr)
	}

	requestBody := gsclientgen.LoginBodyModel{Password: string(encodedPassword)}
	loginResponse, _, err := apiClient.UserLogin(args.email, requestBody, requestIDHeader, loginActivityName, cmdLine)
	if err != nil {
		return result, err
	}

	switch loginResponse.StatusCode {
	case apischema.STATUS_CODE_DATA:
		// successful login
		result.token = loginResponse.Data.Id
		result.email = args.email

		// fetch installation name as alias
		authHeader := "giantswarm " + result.token
		infoResponse, _, infoErr := apiClient.GetInfo(authHeader, requestIDHeader, loginActivityName, cmdLine)
		if infoErr != nil {
			return result, microerror.Mask(infoErr)
		}

		result.alias = infoResponse.General.InstallationName

		if err := config.Config.StoreEndpointAuth(args.apiEndpoint, result.alias, args.email, result.token); err != nil {
			return result, microerror.Mask(err)
		}
		if err := config.Config.SelectEndpoint(args.apiEndpoint); err != nil {
			return result, microerror.Mask(err)
		}

		return result, nil

	case apischema.STATUS_CODE_RESOURCE_INVALID_CREDENTIALS:
		// bad credentials
		return result, microerror.Mask(invalidCredentialsError)
	case apischema.STATUS_CODE_RESOURCE_NOT_FOUND:
		// user unknown or user/password mismatch
		return result, microerror.Mask(invalidCredentialsError)
	case apischema.STATUS_CODE_WRONG_INPUT:
		// empty password
		return result, microerror.Mask(emptyPasswordError)
	default:
		return result, fmt.Errorf("Unhandled response code: %v", loginResponse.StatusCode)
	}
}
