package util

import "github.com/juju/errgo"

var CouldNotSetKubectlClusterError = errgo.New("could not set cluster using 'kubectl config set-cluster'")

// IsCouldNotSetKubectlClusterError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlClusterError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlClusterError
}

var CouldNotSetKubectlCredentialsError = errgo.New("could not set credentials using 'kubectl config set-credentials'")

// IsCouldNotSetKubectlCredentialsError asserts CouldNotSetKubectlClusterError.
func IsCouldNotSetKubectlCredentialsError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlCredentialsError
}

var CouldNotSetKubectlContextError = errgo.New("could not set context using 'kubectl config sez-context''")

// IsCouldNotSetKubectlContextError asserts CouldNotSetKubectlContextError.
func IsCouldNotSetKubectlContextError(err error) bool {
	return errgo.Cause(err) == CouldNotSetKubectlContextError
}

var CouldNotUseKubectlContextError = errgo.New("could not apply context using 'kubectl config use-context'")

// IsCouldNotUseKubectlContextError asserts CouldNotUseKubectlContextError.
func IsCouldNotUseKubectlContextError(err error) bool {
	return errgo.Cause(err) == CouldNotUseKubectlContextError
}
