package cluster

import "github.com/giantswarm/microerror"

// unmashalToMapFailedError is used when a YAML cluster definition can't be unmarshalled into map[string]interface{}.
var unmashalToMapFailedError = &microerror.Error{
	Kind: "unmashalToMapFailedError",
	Desc: "Could not unmarshal YAML into a map[string]interface{} structure. Seems like the YAML is invalid.",
}

// IsUnmashalToMapFailed asserts unmashalToMapFailedError.
func IsUnmashalToMapFailed(err error) bool {
	return microerror.Cause(err) == unmashalToMapFailedError
}
