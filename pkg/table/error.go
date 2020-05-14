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

var columnNotFoundError = &microerror.Error{
	Kind: "columnNotFoundError",
}

// IsColumnNotFoundError asserts columnNotFoundError.
func IsColumnNotFoundError(err error) bool {
	return microerror.Cause(err) == columnNotFoundError
}
