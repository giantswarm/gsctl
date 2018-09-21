package util

import "github.com/giantswarm/microerror"

// CouldNotSetKubectlClusterError is used if kubectl config set-cluster could not be executed.
var CouldNotSetKubectlClusterError = microerror.New("could not set cluster using 'kubectl config set-cluster'")

// IsCouldNotSetKubectlClusterError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlClusterError(err error) bool {
	return microerror.Cause(err) == CouldNotSetKubectlClusterError
}

// CouldNotSetKubectlCredentialsError is used when kubectl config set-credentials could not be executed.
var CouldNotSetKubectlCredentialsError = microerror.New("could not set credentials using 'kubectl config set-credentials'")

// IsCouldNotSetKubectlCredentialsError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlCredentialsError(err error) bool {
	return microerror.Cause(err) == CouldNotSetKubectlCredentialsError
}

// CouldNotSetKubectlContextError is used when kubectl config set-context could not be executed.
var CouldNotSetKubectlContextError = microerror.New("could not set context using 'kubectl config set-context''")

// IsCouldNotSetKubectlContextError asserts CouldNotSetKubectlContextError.
func IsCouldNotSetKubectlContextError(err error) bool {
	return microerror.Cause(err) == CouldNotSetKubectlContextError
}

// CouldNotUseKubectlContextError is used when kubectl config use-context could not be executed.
var CouldNotUseKubectlContextError = microerror.New("could not apply context using 'kubectl config use-context'")

// IsCouldNotUseKubectlContextError asserts CouldNotUseKubectlContextError.
func IsCouldNotUseKubectlContextError(err error) bool {
	return microerror.Cause(err) == CouldNotUseKubectlContextError
}

// InvalidDurationStringError is used when a duration string given by the user could not be parsed.
var InvalidDurationStringError = &microerror.Error{
	Kind: "InvalidDurationStringError",
}

// IsInvalidDurationStringError asserts InvalidDurationStringError.
func IsInvalidDurationStringError(err error) bool {
	return microerror.Cause(err) == InvalidDurationStringError
}

// durationExceededError is thrown when a duration value is larger than can be represented internally
var durationExceededError = &microerror.Error{
	Kind: "durationExceededError",
}

// IsDurationExceededError asserts durationExceededError
func IsDurationExceededError(err error) bool {
	return microerror.Cause(err) == durationExceededError
}
