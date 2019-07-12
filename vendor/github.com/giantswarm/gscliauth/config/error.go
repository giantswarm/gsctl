package config

import "github.com/giantswarm/microerror"

// endpointNotDefinedError means the user tries to use an endpoint that is not defined.
var endpointNotDefinedError = &microerror.Error{
	Kind: "endpointNotDefinedError",
}

// IsEndpointNotDefinedError asserts endpointNotDefinedError.
func IsEndpointNotDefinedError(err error) bool {
	return microerror.Cause(err) == endpointNotDefinedError
}

// noEndpointSelectedError means no endpoint is currently selected.
var noEndpointSelectedError = &microerror.Error{
	Kind: "noEndpointSelectedError",
}

// IsNoEndpointSelectedError asserts noEndpointSelectedError.
func IsNoEndpointSelectedError(err error) bool {
	return microerror.Cause(err) == noEndpointSelectedError
}

// endpointProviderIsImmuttableError means no endpoint is currently selected.
var endpointProviderIsImmuttableError = &microerror.Error{
	Kind: "endpointProviderIsImmuttableError",
}

// IsEndpointProviderIsImmuttableError asserts endpointProviderIsImmuttableError.
func IsEndpointProviderIsImmuttableError(err error) bool {
	return microerror.Cause(err) == endpointProviderIsImmuttableError
}

// aliasMustBeUniqueError should be used if the user tries to add an alias to
// an endpoint, but the alias is already in use.
var aliasMustBeUniqueError = &microerror.Error{
	Kind: "aliasMustBeUniqueError",
}

// IsAliasMustBeUniqueError asserts aliasMustBeUniqueError.
func IsAliasMustBeUniqueError(err error) bool {
	return microerror.Cause(err) == aliasMustBeUniqueError
}

// credentialsRequiredError means an attempt to store incomplete credentials in the config.
var credentialsRequiredError = &microerror.Error{
	Kind: "credentialsRequiredError",
	Desc: "email, password, or token must not be empty",
}

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

// unableToRefreshToken indicates that we were not able to get a new access token
// when we attempted to do so.
var unableToRefreshTokenError = &microerror.Error{
	Kind: "IsUnableToRefreshTokenError",
}

// IsUnableToRefreshTokenErrorr asserts unableToRefreshTokenError.
func IsUnableToRefreshTokenErrorr(err error) bool {
	return microerror.Cause(err) == unableToRefreshTokenError
}
