package pkce

import "github.com/giantswarm/microerror"

var authorizationError = &microerror.Error{
	Kind: "authorizationError",
}

// IsAuthorizationError asserts authorizationError.
func IsAuthorizationError(err error) bool {
	return microerror.Cause(err) == authorizationError
}

// To be used when a token's signature or syntax is invalid
var tokenInvalidError = &microerror.Error{
	Kind: "tokenInvalidError",
}

// IsTokenInvalidError asserts tokenInvalidError.
func IsTokenInvalidError(err error) bool {
	return microerror.Cause(err) == tokenInvalidError
}

// To be used when a token's iat claim (issued at) is bad
var tokenIssuedAtError = &microerror.Error{
	Kind: "tokenIssuedAtError",
}

// IsTokenIssuedAtError asserts tokenIssuedAtError.
func IsTokenIssuedAtError(err error) bool {
	return microerror.Cause(err) == tokenIssuedAtError
}
