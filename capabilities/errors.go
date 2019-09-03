package capabilities

import (
	"github.com/Masterminds/semver"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var couldNotFetchFeatures = &microerror.Error{
	Kind: "couldNotFetchFeatures",
}

// IsCouldNotFetchFeatures asserts couldNotFetchFeatures.
func IsCouldNotFetchFeatures(err error) bool {
	return microerror.Cause(err) == couldNotFetchFeatures
}

var couldNotInitializeCapabilities = &microerror.Error{
	Kind: "couldNotInitializeCapabilities",
}

// IsCouldNotInitializeCapabilities asserts couldNotInitializeCapabilities.
func IsCouldNotInitializeCapabilities(err error) bool {
	return microerror.Cause(err) == couldNotInitializeCapabilities
}

// IsInvalidSemVer asserts semver.ErrInvalidSemVer, as semver unfortunately
// does not provide a matcher.
func IsInvalidSemVer(err error) bool {
	return microerror.Cause(err) == semver.ErrInvalidSemVer
}
