package commands

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	microerror "github.com/giantswarm/microkit/error"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
)

const (
	loginActivityName = "login"

	// errors
	errTokenArgumentNotApplicable = "token argument cannot be used here"
	errNoEmailArgumentGiven       = "no email argument given"
	errInvalidCredentials         = "invalid credentials submitted"
	errEmptyPassword              = "password must not be empty"
)

var (
	// cmdPassword is the password given via command line flag
	cmdPassword string

	// email address passed as a positional argument
	cmdEmail string

	// LoginCommand is the "login" CLI command
	LoginCommand = &cobra.Command{
		Use:    "login <email>",
		Short:  "Sign in as a user",
		Long:   `Sign in with email address and password. Password has to be entered interactively or given as -p flag.`,
		PreRun: loginValidationOutput,
		Run:    loginOutput,
	}
)

type loginResult struct {
	// apiEndpoint is the API endpoint the user has been logged in to
	apiEndpoint string
	// loggedOutBefore is true if the user has been logged out from a previous session
	loggedOutBefore bool
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
		apiEndpoint: cmdAPIEndpoint,
		email:       cmdEmail,
		password:    cmdPassword,
		verbose:     cmdVerbose,
	}
}

func init() {
	LoginCommand.Flags().StringVarP(&cmdPassword, "password", "p", "", "Password. If not given, will be prompted interactively.")
	RootCommand.AddCommand(LoginCommand)
}

// loginValidationOutput runs our pre-checks.
// If an error occurred, it prints the error info and exits with non-zero code.
func loginValidationOutput(cmd *cobra.Command, positionalArgs []string) {
	err := loginValidation(positionalArgs)

	if err != nil {
		var headline = ""
		var subtext = ""
		switch err.Error() {
		case "":
			return
		case errNoEmailArgumentGiven:
			headline = "The email argument is required."
			subtext = "Please execute the command as 'gsctl login <email>'. See 'gsctl login --help' for details."
		case errTokenArgumentNotApplicable:
			headline = "The '--auth-token' flag cannot be used with the 'gsctl login' command."
		case errEmptyPassword:
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

// loginValidation does the pre-checks and returns an error in case something's wrong.
func loginValidation(positionalArgs []string) error {
	if len(positionalArgs) >= 1 {
		// set cmdEmail for later use, as cobra doesn't do that for us
		cmdEmail = positionalArgs[0]
	} else {
		return errors.New(errNoEmailArgumentGiven)
	}

	// using auth token flag? The 'login' command is the only exception
	// where we can't accept this argument.
	if cmdToken != "" {
		return errors.New(errTokenArgumentNotApplicable)
	}

	// interactive password prompt
	if cmdPassword == "" {
		fmt.Printf("Password for %s: ", cmdEmail)
		password, err := gopass.GetPasswd()
		if err != nil {
			return err
		}
		if string(password) == "" {
			return errors.New(errEmptyPassword)
		}
		cmdPassword = string(password)
	}

	return nil
}

// loginOutput executes the login logic and
// prints output and sets the exit code.
func loginOutput(cmd *cobra.Command, args []string) {
	loginArgs := defaultLoginArguments()

	result, err := login(loginArgs)
	if err != nil {
		var headline = ""
		var subtext = ""
		switch err.Error() {
		case "":
			return
		case errEmptyPassword:
			headline = "Empty password submitted"
			subtext = "The API server complains about the password provided."
			subtext += " Please make sure to provide a string with more than white space characters."
		case errInvalidCredentials:
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

	fmt.Println(color.GreenString("You are logged in as %s at %s.",
		result.email, result.apiEndpoint))
}

// login executes the authentication logic.
// If the user was logged in before, a logout is performed first.
func login(args loginArguments) (loginResult, error) {
	result := loginResult{
		apiEndpoint:     args.apiEndpoint,
		email:           args.email,
		loggedOutBefore: false,
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
		return result, microerror.MaskAny(couldNotCreateClientError)
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
		config.Config.Token = result.token
		config.Config.Email = args.email
		config.WriteToFile()

		return result, nil
	case apischema.STATUS_CODE_RESOURCE_INVALID_CREDENTIALS:
		// bad credentials
		return result, errors.New(errInvalidCredentials)
	case apischema.STATUS_CODE_RESOURCE_NOT_FOUND:
		// user unknown or user/password mismatch
		return result, errors.New(errInvalidCredentials)
	case apischema.STATUS_CODE_WRONG_INPUT:
		// empty password
		return result, errors.New(errEmptyPassword)
	default:
		return result, fmt.Errorf("Unhandled response code: %v", loginResponse.StatusCode)
	}
}
