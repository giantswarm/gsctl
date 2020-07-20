package nodepool

import (
	"github.com/giantswarm/microerror"
)

var invalidAvailabilityZones = &microerror.Error{
	Kind: "invalidAvailabilityZones",
}

// IsInvalidAvailabilityZones asserts invalidAvailabilityZones.
func IsInvalidAvailabilityZones(err error) bool {
	return microerror.Cause(err) == invalidAvailabilityZones
}
