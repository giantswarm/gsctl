# Blueprint for a command file

This example shows a scaffold for a fictitious `gsctl verb noun` command.

```go
package commands

// imports:
// - standard library first
// - external dependencies next
// - gsctl sub-packages last
import (
  "fmt"

  "github.com/fatih/color"
  "github.com/spf13/cobra"
  "github.com/giantswarm/microerror"
  "github.com/giantswarm/gscliauth/config"

  "github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/util"
)

// variables. Here we create the command
var (
  // VerbNounCommand performs the "verb noun" function
	VerbNounCommand = &cobra.Command{
		Use:     "noun",
		Short:   "Does something with noun",
		Long:    `Manipulates noun in a way only verb can

This may span multiple lines.`,

    // We use PreRun for general input validation, authentication etc.
    // If something is bad/missing, that function has to exit with a
    // non-zero exit code.
		PreRun:  verbNounPreRunOutput,

    // Run is the function that actually executes what we want to do.
		Run:     verbNounRunOutput,
	}

  // global variable to be assigned by command line flag
  cmdMyFlag string
)

const (
  // verbNounActivityName assigns API requests to named activities
  verbNounActivityName = "verb-noun"
)

// argument struct to pass to our business function and
// to the validation function
type verbNounArguments struct {
	apiEndpoint     string
	authToken       string
	anotherArgument string
}

// function to create arguments based on command line flags and config
func defaultVerbNounArguments() verbNounArguments {
	endpoint := config.Config.ChooseEndpoint(CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, CmdToken)

	return verbNounArguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		anotherArgument:   "",
	}
}

// verbNounResult is used to return a structured result
// from our business function
type verbNounResult struct {
  someAttribute string
}

// Here we populate our cobra command
func init() {
	VerbNounCommand.Flags().StringVarP(&cmdMyFlag, "myflag", "m", "", "Placeholder flag")
	VerbNounCommand.MarkFlagRequired("myflag")

	VerbCommand.AddCommand(VerbNounCommand)
}

// Prints results of our pre-validation
func verbNounPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultVerbNounArguments()
	err := verifyVerbNounPreconditions(args, cmdLineArgs)

  if err == nil {
    return
  }

  // Handles many errors that can occur in validation and execution,
  // e. g. user not logged in.
  HandleCommonErrors(err)

  // From here on we handle errors that can only occur in this command
	headline := ""
	subtext := ""

	switch {
	case IsVerySpecificError(err):
		headline = "Some very specific error occurred."
		subtext = "Something happened that can only happen in this command."
	default:
		headline = err.Error()
	}

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

// Checks if all preconditions are met, before actually executing
// our business function
func verifyVerbNounPreconditions(args verbNounArguments, cmdLineArgs []string) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(NotLoggedInError)
	}
	return nil
}

// verbNounRunOutput executes our business function and displays the result,
// both in case of success or error
func verbNounRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultVerbNounArguments()
	result, err := verbNoun()

	if err != nil {
    HandleCommonErrors(err)

		var headline = ""
		var subtext = ""

		switch {
		case IsVerySpecificError(err):
      headline = "Some very specific error occurred."
  		subtext = "Something happened that can only happen in this command."
		default:
			headline = err.Error()
		}

		// Print error output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	fmt.Println("Success!")
}

// verbNoun performs our actual function. It usually creates an API client,
// configures it, configures an API request and performs it.
// verbNoun performs our actual function. It usually creates an API client,
// configures it, configures an API request and performs it.
func verbNoun(args verbNoundArguments) (verbNounResult, error) {
	result := verbNounResult{}

	// prepare client
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return result, microerror.Mask(CouldNotCreateClientError)
	}

	authHeader := "giantswarm " + args.token
	someResponse, rawResponse, err := apiClient.DoSomething(authHeader,
		requestIDHeader, verbNounActivityName, cmdLine)

	if rawResponse == nil || rawResponse.Response == nil {
		return result, microerror.Mask(NoResponseError)
	}

	// handle request errors
	if err != nil {

		switch rawResponse.StatusCode {
		case http.StatusNotFound:
			return result, microerror.Mask(ClusterNotFoundError)
		case http.StatusUnauthorized:
			return result, microerror.Mask(NotAuthorizedError)
		case http.StatusForbidden:
			return result, microerror.Mask(AccessForbiddenError)
		}

		if rawResponse.StatusCode >= 500 {
			return result, microerror.Maskf(InternalServerError, err.Error())
		}

		return result, microerror.Mask(err)
	}

	// populate result base on some response information etc.
	result.someAttribute = someResponse.someValue

	return result, nil
}
```
