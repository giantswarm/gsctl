package client

import (
	"net"

	"github.com/juju/errgo"
)

// IsTimeoutError is a matcher for the golang-internal net timeout error
func IsTimeoutError(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}

// endpointNotSpecifiedError is used in an attempt to create a client without endpoint
var endpointNotSpecifiedError = errgo.New("no endpoint has been specified")

// IsEndpointNotSpecifiedError asserts endpointNotSpecifiedError.
func IsEndpointNotSpecifiedError(err error) bool {
	return errgo.Cause(err) == endpointNotSpecifiedError
}
