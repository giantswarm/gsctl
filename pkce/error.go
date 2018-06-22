package pkce

import "github.com/juju/errgo"

var authorizationError = errgo.New("authorization error")

// IsAuthorizationError asserts authorizationError.
func IsAuthorizationError(err error) bool {
	return errgo.Cause(err) == authorizationError
}
