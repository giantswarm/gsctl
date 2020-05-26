package cluster

import (
	"github.com/giantswarm/microerror"
)

var revertHAMasterNotAllowedError = &microerror.Error{
	Kind: "revertHAMasterNotAllowedError",
}

// IsRevertHAMasterNotAllowed asserts revertHAMasterNotAllowedError.
func IsRevertHAMasterNotAllowed(err error) bool {
	return microerror.Cause(err) == revertHAMasterNotAllowedError
}

var haMastersNotSupportedError = &microerror.Error{
	Kind: "haMastersNotSupportedError",
}

// IsHAMastersNotSupported asserts haMastersNotSupportedError.
func IsHAMastersNotSupported(err error) bool {
	return microerror.Cause(err) == haMastersNotSupportedError
}
