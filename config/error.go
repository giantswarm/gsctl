package config

import "github.com/giantswarm/microerror"

// endpointNotDefinedError means the user tries to use an endpoint that is not defined
var endpointNotDefinedError = &microerror.Error{
	Kind: "endpointNotDefinedError",
}

// IsEndpointNotDefinedError asserts endpointNotDefinedError.
func IsEndpointNotDefinedError(err error) bool {
	return microerror.Cause(err) == endpointNotDefinedError
}

// aliasMustBeUniqueError should be used if the user tries to add an alias to
// an endpoint, but the alias is already in use
var aliasMustBeUniqueError = &microerror.Error{
	Kind: "aliasMustBeUniqueError",
}

// IsAliasMustBeUniqueError asserts aliasMustBeUniqueError.
func IsAliasMustBeUniqueError(err error) bool {
	return microerror.Cause(err) == aliasMustBeUniqueError
}

// credentialsRequiredError means an attempt to store incomplete credentials in the config
var credentialsRequiredError = microerror.New("email, password, or token must not be empty")

// IsCredentialsRequiredError asserts credentialsRequiredError.
func IsCredentialsRequiredError(err error) bool {
	return microerror.Cause(err) == credentialsRequiredError
}

var garbageCollectionFailedError = &microerror.Error{
	Kind: "garbageCollectionFailedError",
}

// IsGarbageCollectionFailedError asserts garbageCollectionFailedError.
func IsGarbageCollectionFailedError(err error) bool {
	return microerror.Cause(err) == garbageCollectionFailedError
}

var garbageCollectionPartiallyFailedError = &microerror.Error{
	Kind: "garbageCollectionPartiallyFailedError",
}

// IsGarbageCollectionPartiallyFailedError asserts garbageCollectionPartiallyFailedError.
func IsGarbageCollectionPartiallyFailedError(err error) bool {
	return microerror.Cause(err) == garbageCollectionPartiallyFailedError
}
