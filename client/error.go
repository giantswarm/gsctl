package client

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client/clienterror"
)

// clientV2NotInitializedError is used when the new client hasn't been initialized.
var clientV2NotInitializedError = &microerror.Error{
	Kind: "clientV2NotInitializedError",
}

// IsClientV2NotInitializedError asserts clientV2NotInitializedError.
func IsClientV2NotInitializedError(err error) bool {
	return microerror.Cause(err) == clientV2NotInitializedError
}

// endpointInvalidError is used if an endpoint string is not a valid URL.
var endpointInvalidError = &microerror.Error{
	Kind: "endpointInvalidError",
}

// IsEndpointInvalidError asserts endpointInvalidError.
func IsEndpointInvalidError(err error) bool {
	return microerror.Cause(err) == endpointInvalidError
}

// endpointNotSpecifiedError is used in an attempt to create a client without endpoint.
var endpointNotSpecifiedError = &microerror.Error{
	Kind: "endpointNotSpecifiedError",
}

// IsEndpointNotSpecifiedError asserts endpointNotSpecifiedError.
func IsEndpointNotSpecifiedError(err error) bool {
	return microerror.Cause(err) == endpointNotSpecifiedError
}

// NotAuthorizedError is used when an API request got a 401 response.
var NotAuthorizedError = &microerror.Error{
	Kind: "NotAuthorizedError",
}

// IsNotAuthorizedError asserts NotAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return microerror.Cause(err) == NotAuthorizedError
}

// HandleErrors handles the errors known to this package.
// Handling normally means printing a user-readable error message
// and exiting with code 1. If the given error is not recognized,
// the function returns without action.
func HandleErrors(err error) {

	var headline = ""
	var subtext = ""

	// V2 client error handling
	if convertedErr, ok := microerror.Cause(err).(*clienterror.APIError); ok {
		headline = convertedErr.ErrorMessage
		subtext = convertedErr.ErrorDetails
	} else if convertedErr, ok := err.(*clienterror.APIError); ok {
		headline = convertedErr.ErrorMessage
		subtext = convertedErr.ErrorDetails
	} else if IsEndpointNotSpecifiedError(err) {
		// legacy client error handling
		headline = "No endpoint has been specified."
		subtext = "Please use the '-e|--endpoint' flag or select an endpoint using 'gsctl select endpoint'."
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
