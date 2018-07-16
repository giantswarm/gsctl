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
