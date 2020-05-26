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
