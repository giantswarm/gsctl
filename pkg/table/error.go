package table

import (
	"github.com/giantswarm/microerror"
)

var objectNotSliceError = &microerror.Error{
	Kind: "objectNotSliceError",
}

// IsObjectNotSliceError asserts objectNotSliceError.
func IsObjectNotSliceError(err error) bool {
	return microerror.Cause(err) == objectNotSliceError
}

var fieldNotFoundError = &microerror.Error{
	Kind: "fieldNotFoundError",
}

// IsFieldNotFoundError asserts fieldNotFoundError.
func IsFieldNotFoundError(err error) bool {
	return microerror.Cause(err) == fieldNotFoundError
}

var multipleFieldsMatchingError = &microerror.Error{
	Kind: "multipleFieldsMatchingError",
}

// IsMultipleFieldsMatchingError asserts multipleFieldsMatchingError.
func IsMultipleFieldsMatchingError(err error) bool {
	return microerror.Cause(err) == multipleFieldsMatchingError
}
