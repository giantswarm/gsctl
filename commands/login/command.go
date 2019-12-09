package login

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gscliauth/oidc"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

const (
	loginActivityName = "login"
)

var (
	// cmdPassword is the password given via command line flag
	cmdPassword string

	// email address passed as a positional argument
	cmdEmail string

	// cmdSSO is the bool that triggers login via SSO.
	cmdSSO bool

	// Command is the "login" CLI command
	Command = &cobra.Command{
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

func init() {
	Command.Flags().StringVarP(&cmdPassword, "password", "p", "", "Password. If not given, will be prompted interactively.")
	Command.Flags().BoolVarP(&cmdSSO, "sso", "", false, "Authenticate using Single Sign On through our identity provider.")
	Command.Flags().MarkHidden("sso")
}

// Arguments is the argument struct for the business function.
// Note: the absence of 'token', which is available in all other commands,
// is by design.
type Arguments struct {
	apiEndpoint string
	email       string
	password    string
	verbose     bool
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)

	return Arguments{
		apiEndpoint: endpoint,
		email:       cmdEmail,
		password:    cmdPassword,
		verbose:     flags.Verbose,
	}
}

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
	// provider is the provider of the installation
	provider string
	// token is the new session token received
	token string
	// numEndpointsBefore is the number of endpoints before login
	numEndpointsBefore int
	// numEndpointsAfter is the number of endpoints after login
	numEndpointsAfter int
}

// loginPreRunOutput runs our pre-checks.
// If an error occurred, it prints the error info and exits with non-zero code.
func loginPreRunOutput(cmd *cobra.Command, positionalArgs []string) {
	err := verifyLoginPreconditions(positionalArgs)

	if err == nil {
		return
	}

	var headline = ""
	var subtext = ""

	switch {
	case errors.IsNoEmailArgumentGivenError(err):
		headline = "The email argument is required."
		subtext = "Please execute the command as 'gsctl login <email>'. See 'gsctl login --help' for details."
	case errors.IsTokenArgumentNotApplicableError(err):
		headline = "The '--auth-token' flag cannot be used with the 'gsctl login' command."
	case errors.IsPasswordArgumentNotApplicableError(err):
		headline = "The '--password' flag cannot be used with the 'gsctl login --sso' command."
	case errors.IsEmptyPasswordError(err):
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

// verifyLoginPreconditions does the pre-checks and returns an error in case something's wrong.
func verifyLoginPreconditions(positionalArgs []string) error {
	args := collectArguments()

	// using auth token flag? The 'login' command is the only exception
	// where we can't accept this argument.
	if flags.Token != "" {
		return microerror.Mask(errors.TokenArgumentNotApplicableError)
	}

	if cmdSSO {
		if cmdPassword != "" {
			return microerror.Mask(errors.PasswordArgumentNotApplicableError)
		}
	} else {
		if len(positionalArgs) >= 1 {
			// set cmdEmail for later use, as cobra doesn't do that for us
			cmdEmail = positionalArgs[0]
		} else {
			return microerror.Mask(errors.NoEmailArgumentGivenError)
		}

		// interactive password prompt
		if cmdPassword == "" {
			fmt.Printf("Password for %s on %s: ", color.CyanString(cmdEmail), color.CyanString(args.apiEndpoint))
			password, err := gopass.GetPasswd()
			if err != nil {
				return err
			}
			if string(password) == "" {
				return microerror.Mask(errors.EmptyPasswordError)
			}
			cmdPassword = string(password)
		}
	}

	return nil
}

func login(loginArgs Arguments) (loginResult, error) {
	var result loginResult
	var err error
	if cmdSSO {
		result, err = loginSSO(loginArgs)
	} else {
		result, err = loginGiantSwarm(loginArgs)
	}

	return result, err
}

// loginRunOutput executes the login logic and
// prints output and sets the exit code.
func loginRunOutput(cmd *cobra.Command, args []string) {
	loginArgs := collectArguments()

	result, err := login(loginArgs)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		var headline = ""
		var subtext = ""
		switch {
		case errors.IsEmptyPasswordError(err):
			headline = "Empty password submitted"
			subtext = "The API server complains about the password provided."
			subtext += " Please make sure to provide a string with more than white space characters."
		case errors.IsInvalidCredentialsError(err):
			headline = "Bad password or email address"
			subtext = fmt.Sprintf("Could not log you in to %s.", color.CyanString(loginArgs.apiEndpoint))
			subtext += " The email or the password provided (or both) was incorrect."
		case errors.IsUserAccountInactiveError(err):
			headline = "User account has expired or is deactivated"
			subtext = "Please contact the Giant Swarm support team."
		case config.IsAliasMustBeUniqueError(err):
			headline = "Alias is already in use for a different endpoint"
			subtext = fmt.Sprintf("The alias '%s' is already used for an endpoint in your configuration.\n", result.alias)
			subtext += "Please edit your configuration file manually to delete the alias or endpoint."
		case oidc.IsTokenIssuedAtError(err):
			headline = "Token created in the future?"
			subtext = "It appears as if your system time is behind the actual time. Please adjust the time and make sure\n"
			subtext += "that it is automatically synchronized with a time service. Otherwise SSO login does not work."
		case errors.IsSSOError(err):
			headline = "Something went wrong during SSO"
			subtext = err.Error()
			subtext += "\nPlease contact the Giant Swarm support team or try the command again later."
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
		if result.alias != "" {
			fmt.Printf("Endpoint selected: %s (%s)\n", result.apiEndpoint, result.alias)
		} else {
			fmt.Printf("Endpoint selected: %s\n", result.apiEndpoint)
		}
	}

	fmt.Println(color.GreenString("You are logged in as %s at %s.",
		result.email, result.apiEndpoint))

	// we only want this extra hint on endpoint switching if
	// - at least two endpoints in total
	// - an endpoint has been just added
	// - the new endpoint has an alias
	if result.numEndpointsAfter > result.numEndpointsBefore && result.numEndpointsAfter > 1 && result.alias != "" {
		fmt.Println()
		fmt.Println(color.GreenString("To switch back to this endpoint, you can use this command:\n"))
		fmt.Println(color.YellowString("    gsctl select endpoint %s\n", result.alias))
	}
}

// getInstallationInfo creates a giantswarm API client and tries to fetch the info endpoint.
// If it succeeds it returns the alias for that endpoint.
func getInstallationInfo(apiEndpoint string, scheme string, accessToken string) (*models.V4InfoResponseGeneral, error) {
	// Create an API client.
	authHeaderGetter := func() (string, error) {
		return scheme + " " + accessToken, nil
	}

	clientConfig := &client.Configuration{
		Endpoint:         apiEndpoint,
		Timeout:          10 * time.Second,
		UserAgent:        config.UserAgent(),
		AuthHeaderGetter: authHeaderGetter,
	}

	clientWrapper, err := client.New(clientConfig)
	if err != nil {
		return nil, microerror.Maskf(errors.CouldNotCreateClientError, err.Error())
	}

	// Fetch installation name as alias.
	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = loginActivityName
	infoResponse, err := clientWrapper.GetInfo(auxParams)
	if err != nil {
		return nil, err
	}

	return infoResponse.Payload.General, nil
}
