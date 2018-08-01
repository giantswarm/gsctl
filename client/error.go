package client

import "github.com/juju/errgo"

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

// notAuthorizedError is used when an API request got a 401 response.
var notAuthorizedError = &microerror.Error{
	Kind: "notAuthorizedError",
}

// IsNotAuthorizedError asserts notAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return microerror.Cause(err) == notAuthorizedError
}
