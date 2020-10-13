package releaseinfo

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var internalServerError = &microerror.Error{
	Kind: "internalServerError",
}

// IsInternalServerError asserts internalServerError.
func IsInternalServerError(err error) bool {
	return microerror.Cause(err) == internalServerError
}

var notAuthorizedError = &microerror.Error{
	Kind: "notAuthorizedError",
}

// IsNotAuthorized asserts notAuthorizedError.
func IsNotAuthorized(err error) bool {
	return microerror.Cause(err) == notAuthorizedError
}

var versionNotFoundError = &microerror.Error{
	Kind: "versionNotFoundError",
}

// IsVersionNotFound asserts versionNotFoundError.
func IsVersionNotFound(err error) bool {
	return microerror.Cause(err) == versionNotFoundError
}

var componentNotFoundError = &microerror.Error{
	Kind: "componentNotFoundError",
}

// IsComponentNotFound asserts componentNotFoundError.
func IsComponentNotFound(err error) bool {
	return microerror.Cause(err) == componentNotFoundError
}
