package pkce

import "github.com/juju/errgo"

var authorizationError = errgo.New("authorization error")

// IsAuthorizationError asserts authorizationError.
func IsAuthorizationError(err error) bool {
	return errgo.Cause(err) == authorizationError
}

// To be used when a token's signature or syntax is invalid
var tokenInvalidError = errgo.New("token invalid")

// IsTokenInvalidError asserts tokenInvalidError.
func IsTokenInvalidError(err error) bool {
	return errgo.Cause(err) == tokenInvalidError
}

// To be used when a token's iat claim (issued at) is bad
var tokenIssuedAtError = errgo.New("token issued at invalid time")

// IsTokenIssuedAtError asserts tokenIssuedAtError.
func IsTokenIssuedAtError(err error) bool {
	return errgo.Cause(err) == tokenIssuedAtError
}
