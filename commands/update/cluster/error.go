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

var onlyV5SupportedError = &microerror.Error{
	Kind: "onlyV5SupportedError",
}

// IsOnlyV5Supported asserts onlyV5SupportedError.
func IsOnlyV5Supported(err error) bool {
	return microerror.Cause(err) == onlyV5SupportedError
}
