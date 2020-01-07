package webui

import "github.com/giantswarm/microerror"

var unsupportedHostNameError = &microerror.Error{
	Kind: "unsupportedHostNameError",
	Desc: "The host name is of a format that we do not support",
}

// IsUnsupportedHostName asserts unsupportedHostNameError.
func IsUnsupportedHostName(err error) bool {
	return microerror.Cause(err) == unsupportedHostNameError
}

var missingArgumentError = &microerror.Error{
	Kind: "missingArgumentError",
}

// IsMissingArgument asserts missingArgumentError.
func IsMissingArgument(err error) bool {
	return microerror.Cause(err) == missingArgumentError
}
