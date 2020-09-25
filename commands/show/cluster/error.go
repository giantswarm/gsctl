package cluster

import "github.com/giantswarm/microerror"

var unknownProviderError = &microerror.Error{
	Kind: "unknownProviderError",
}

// IsUnknownProvider asserts unknownProviderError.
func IsUnknownProvider(err error) bool {
	return microerror.Cause(err) == unknownProviderError
}

var providerInfoCorruptError = &microerror.Error{
	Kind: "providerInfoCorruptError",
}

// IsProviderInfoCorrupt asserts providerInfoCorruptError.
func IsProviderInfoCorrupt(err error) bool {
	return microerror.Cause(err) == providerInfoCorruptError
}
