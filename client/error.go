package client

import "github.com/juju/errgo"

// endpointInvalidError is used if an endpoint string is not a valid URL.
var endpointInvalidError = errgo.New("not a valid endpoint URL")

// IsEndpointInvalidError asserts endpointInvalidError.
func IsEndpointInvalidError(err error) bool {
	return errgo.Cause(err) == endpointInvalidError
}

// endpointNotSpecifiedError is used in an attempt to create a client without endpoint.
var endpointNotSpecifiedError = errgo.New("no endpoint has been specified")

// IsEndpointNotSpecifiedError asserts endpointNotSpecifiedError.
func IsEndpointNotSpecifiedError(err error) bool {
	return errgo.Cause(err) == endpointNotSpecifiedError
}

// notAuthorizedError is used when an API request got a 401 response.
var notAuthorizedError = errgo.New("not authorized")

// IsNotAuthorizedError asserts notAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return errgo.Cause(err) == notAuthorizedError
}
