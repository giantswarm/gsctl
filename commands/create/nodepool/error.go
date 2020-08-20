package nodepool

import (
	"github.com/giantswarm/microerror"
)

var invalidAvailabilityZonesError = &microerror.Error{
	Kind: "invalidAvailabilityZonesError",
}

// IsInvalidAvailabilityZones asserts invalidAvailabilityZonesError.
func IsInvalidAvailabilityZones(err error) bool {
	return microerror.Cause(err) == invalidAvailabilityZonesError
}
