package client

import "github.com/juju/errgo"

// endpointNotSpecifiedError is used in an attempt to create a client without endpoint
var endpointNotSpecifiedError = errgo.New("no endpoint has been specified")

// IsEndpointNotSpecifiedError asserts endpointNotSpecifiedError.
func IsEndpointNotSpecifiedError(err error) bool {
	return errgo.Cause(err) == endpointNotSpecifiedError
}
