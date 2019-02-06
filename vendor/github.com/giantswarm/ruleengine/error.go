package ruleengine

import (
	"github.com/giantswarm/microerror"
)

var validationError = &microerror.Error{
	Kind: "validationError",
}

// IsValidation asserts validationError.
func IsValidation(err error) bool {
	return microerror.Cause(err) == validationError
}
