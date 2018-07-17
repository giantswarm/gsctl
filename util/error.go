package util

import "github.com/juju/errgo"

// CouldNotSetKubectlClusterError is used if kubectl config set-cluster could not be executed.
var CouldNotSetKubectlClusterError = errgo.New("could not set cluster using 'kubectl config set-cluster'")

// IsCouldNotSetKubectlClusterError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlClusterError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlClusterError
}

// CouldNotSetKubectlCredentialsError is used when kubectl config set-credentials could not be executed.
var CouldNotSetKubectlCredentialsError = errgo.New("could not set credentials using 'kubectl config set-credentials'")

// IsCouldNotSetKubectlCredentialsError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlCredentialsError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlCredentialsError
}

// CouldNotSetKubectlContextError is used when kubectl config set-context could not be executed.
var CouldNotSetKubectlContextError = errgo.New("could not set context using 'kubectl config set-context''")

// IsCouldNotSetKubectlContextError asserts CouldNotSetKubectlContextError.
func IsCouldNotSetKubectlContextError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlContextError
}

// CouldNotUseKubectlContextError is used when kubectl config use-context could not be executed.
var CouldNotUseKubectlContextError = errgo.New("could not apply context using 'kubectl config use-context'")

// IsCouldNotUseKubectlContextError asserts CouldNotUseKubectlContextError.
func IsCouldNotUseKubectlContextError(err error) bool {
	return errgo.Cause(err) == CouldNotUseKubectlContextError
}

// InvalidDurationStringError is used when a duration string given by the user could not be parsed.
var InvalidDurationStringError = errgo.New("could not parse duration string")

// IsInvalidDurationStringError asserts InvalidDurationStringError.
func IsInvalidDurationStringError(err error) bool {
	return errgo.Cause(err) == InvalidDurationStringError
}
