package errors

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/oidc"
)

// HandleCommonErrors is a common function to handle certain errors happening in
// more than one command. If the error given is handled by the function, it
// prints according text for the end user and exits the process.
// If the error is not recognized, we simply return.
//
func HandleCommonErrors(err error) {

	var headline = ""
	var subtext = ""

	// V2 client error handling
	if convertedErr, ok := microerror.Cause(err).(*clienterror.APIError); ok {
		headline = convertedErr.ErrorMessage
		subtext = convertedErr.ErrorDetails
	} else if convertedErr, ok := err.(*clienterror.APIError); ok {
		headline = convertedErr.ErrorMessage
		subtext = convertedErr.ErrorDetails
	} else {
		// legacy client error handling
		switch {
		case oidc.IsAuthorizationError(err):
			headline = "Unauthorized"
			subtext = "Something went wrong during a OIDC operation: " + err.Error() + "\n"
			subtext += "Please try logging in again."
		case oidc.IsRefreshError(err):
			headline = "Unable to refresh your SSO token."
			subtext = err.Error() + "\n"
			subtext += "Please try loging in again using: gsctl login --sso"
		case IsNotLoggedInError(err):
			headline = "You are not logged in."
			subtext = "Use 'gsctl login' to login or '--auth-token' to pass a valid auth token."
		case IsAccessForbiddenError(err):
			// TODO: remove once the legacy client is no longer used
			headline = "Access Forbidden"
			subtext = "The client has been denied access to the API endpoint with an HTTP status of 403.\n"
			subtext += "Please make sure that you are in the right network or VPN. Once that is verified,\n"
			subtext += "check back with Giant Swarm support that your network is permitted access."
		case IsEmptyPasswordError(err):
			headline = "Empty password submitted"
			subtext = "The API server complains about the password provided."
			subtext += " Please make sure to provide a string with more than white space characters."
		case IsClusterIDMissingError(err):
			headline = "No cluster ID specified."
			subtext = "Please specify a cluster ID. Use --help for details."
		case IsCouldNotCreateClientError(err):
			headline = "Failed to create API client."
			subtext = "Details: " + err.Error()
		case IsNotAuthorizedError(err):
			// TODO: remove once the legacy client is no longer used
			headline = "You are not authorized for this action."
			subtext = "Please check whether you are logged in with the right credentials using 'gsctl info'."
		case IsInternalServerError(err):
			headline = "An internal error occurred."
			subtext = "Please try again in a few minutes. If that does not success, please inform the Giant Swarm support team."
		case IsNoResponseError(err):
			headline = "The API didn't send a response."
			subtext = "Please check your connection using 'gsctl ping'. If your connection is fine,\n"
			subtext += "please try again in a few moments."
		case IsUnknownError(err):
			headline = "An error occurred."
			subtext = "Please notify the Giant Swarm support team, or try the command again in a few moments.\n"
			subtext += fmt.Sprintf("Details: %s", err.Error())
		}

	}

	if headline == "" {
		return
	}

	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}
