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

// invalidV5DefinitionYAMLError is used when the YAML definition can't be parsed as valid v5.
var invalidV5DefinitionYAMLError = &microerror.Error{
	Kind: "invalidV5DefinitionYAMLError",
}

// IsInvalidV5DefinitionYAML asserts invalidV4DefinitionYAMLError.
func IsInvalidV5DefinitionYAML(err error) bool {
	return microerror.Cause(err) == invalidV5DefinitionYAMLError
}

// invalidDefinitionYAMLError is used when the YAML definition can't be parsed as any valid cluster definition.
var invalidDefinitionYAMLError = &microerror.Error{
	Kind: "invalidDefinitionYAMLError",
}

// IsInvalidDefinitionYAML asserts invalidDefinitionYAMLError.
func IsInvalidDefinitionYAML(err error) bool {
	return microerror.Cause(err) == invalidDefinitionYAMLError
}

var haMastersNotSupportedError = &microerror.Error{
	Kind: "haMastersNotSupportedError",
}

// IsHAMastersNotSupported asserts haMastersNotSupportedError.
func IsHAMastersNotSupported(err error) bool {
	return microerror.Cause(err) == haMastersNotSupportedError
}

var mustProvideSingleMasterTypeError = &microerror.Error{
	Kind: "mustProvideSingleMasterTypeError",
}

// IsMustProvideSingleMasterType asserts mustProvideSingleMasterTypeError.
func IsMustProvideSingleMasterType(err error) bool {
	return microerror.Cause(err) == mustProvideSingleMasterTypeError
}
