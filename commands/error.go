package commands

import "github.com/juju/errgo"

// Common errors and matcher functions for the "commands" package.

// unknownError should only be used if there is really no way to
// specify the error any further. Note that there is a more specific
// internalServerError.
var unknownError = errgo.New("unknown error")

// IsUnknownError asserts unknownError.
func IsUnknownError(err error) bool {
	return errgo.Cause(err) == unknownError
}

// couldNotCreateClientError means that a client could not be created
var couldNotCreateClientError = errgo.New("could not create client")

// IsCouldNotCreateClientError asserts couldNotCreateClientError.
func IsCouldNotCreateClientError(err error) bool {
	return errgo.Cause(err) == couldNotCreateClientError
}

// notLoggedInError means that the user is currently not authenticated
var notLoggedInError = errgo.New("user not logged in")

// IsNotLoggedInError asserts notLoggedInError.
func IsNotLoggedInError(err error) bool {
	return errgo.Cause(err) == notLoggedInError
}

// commandAbortedError means that the user has aborted a command or input
var commandAbortedError = errgo.New("user has not confirmed or aborted execution")

// IsCommandAbortedError asserts commandAbortedError
func IsCommandAbortedError(err error) bool {
	return errgo.Cause(err) == commandAbortedError
}

// conflictingFlagsError means that the user combined command line options
// that are incompatible
var conflictingFlagsError = errgo.New("conflicting flags used")

// IsConflictingFlagsError asserts conflictingFlagsError.
func IsConflictingFlagsError(err error) bool {
	return errgo.Cause(err) == conflictingFlagsError
}

// desiredEqualsCurrentStateError means that the user described a desired
// state which is equal to the current state.
var desiredEqualsCurrentStateError = errgo.New("desired state equals current state")

// IsDesiredEqualsCurrentStateError asserts desiredEqualsCurrentStateError.
func IsDesiredEqualsCurrentStateError(err error) bool {
	return errgo.Cause(err) == desiredEqualsCurrentStateError
}

// clusterIDMissingError means a required cluster ID has not been given as input
var clusterIDMissingError = errgo.New("cluster ID not specified")

// IsClusterIDMissingError asserts clusterIDMissingError.
func IsClusterIDMissingError(err error) bool {
	return errgo.Cause(err) == clusterIDMissingError
}

// clusterNotFoundError means that a given cluster does not exist
var clusterNotFoundError = errgo.New("the cluster specified could not be found")

// IsClusterNotFoundError asserts clusterNotFoundError.
func IsClusterNotFoundError(err error) bool {
	return errgo.Cause(err) == clusterNotFoundError
}

// internalServerError should only be used in case of server communication
// being responded to with a response status >= 500.
// See also: unknownError
var internalServerError = errgo.New("an internal server error occurred")

// IsInternalServerError asserts internalServerError.
func IsInternalServerError(err error) bool {
	return errgo.Cause(err) == internalServerError
}

// notAuthorizedError means that an API action could not be performed due to
// an authorization problem (usually a HTTP 401 error)
var notAuthorizedError = errgo.New("not authorized")

// IsNotAuthorizedError asserts notAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return errgo.Cause(err) == notAuthorizedError
}

// Errors for cluster creation

// numWorkerNodesMissingError means that the user has not specified how many
// worker nodes a new cluster should have
var numWorkerNodesMissingError = errgo.New("number of workers not specified")

// IsNumWorkerNodesMissingError asserts numWorkerNodesMissingError.
func IsNumWorkerNodesMissingError(err error) bool {
	return errgo.Cause(err) == numWorkerNodesMissingError
}

// notEnoughWorkerNodesError means that the user has specified a too low
// number of worker nodes for a cluster
var notEnoughWorkerNodesError = errgo.New("not enough workers specified")

// IsNotEnoughWorkerNodesError asserts notEnoughWorkerNodesError.
func IsNotEnoughWorkerNodesError(err error) bool {
	return errgo.Cause(err) == notEnoughWorkerNodesError
}

// notEnoughCPUCoresPerWorkerError means the user did not request enough CPUs
// for the worker nodes
var notEnoughCPUCoresPerWorkerError = errgo.New("not enough CPU cores per worker specified")

// IsNotEnoughCPUCoresPerWorkerError asserts notEnoughCPUCoresPerWorkerError.
func IsNotEnoughCPUCoresPerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughCPUCoresPerWorkerError
}

// notEnoughMemoryPerWorkerError means the user did not request enough RAM
// for the worker nodes
var notEnoughMemoryPerWorkerError = errgo.New("not enough memory per worker specified")

// IsNotEnoughMemoryPerWorkerError asserts notEnoughMemoryPerWorkerError.
func IsNotEnoughMemoryPerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughMemoryPerWorkerError
}

// notEnoughStoragePerWorkerError means the user did not request enough disk space
// for the worker nodes
var notEnoughStoragePerWorkerError = errgo.New("not enough storage per worker specified")

// IsNotEnoughStoragePerWorkerError asserts notEnoughStoragePerWorkerError.
func IsNotEnoughStoragePerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughStoragePerWorkerError
}

// clusterOwnerMissingError means that the user has not specified an owner organization
// for a new cluster
var clusterOwnerMissingError = errgo.New("no cluster owner specified")

// IsClusterOwnerMissingError asserts clusterOwnerMissingError.
func IsClusterOwnerMissingError(err error) bool {
	return errgo.Cause(err) == clusterOwnerMissingError
}

// yamlFileNotReadableError means a YAML file was not readable
var yamlFileNotReadableError = errgo.New("could not read YAML file")

// IsYAMLFileNotReadableError asserts yamlFileNotReadableError.
func IsYAMLFileNotReadableError(err error) bool {
	return errgo.Cause(err) == yamlFileNotReadableError
}

// couldNotCreateJSONRequestBodyError occurs when we could not create a JSON
// request body based on the input we have, so something in out input attributes
// is wrong.
var couldNotCreateJSONRequestBodyError = errgo.New("could not create JSON request body")

// IsCouldNotCreateJSONRequestBodyError asserts couldNotCreateJSONRequestBodyError.
func IsCouldNotCreateJSONRequestBodyError(err error) bool {
	return errgo.Cause(err) == couldNotCreateJSONRequestBodyError
}

// couldNotCreateClusterError should be used if the API call to create a
// cluster has been responded with status >= 400 and none of the other
// more specific errors apply.
var couldNotCreateClusterError = errgo.New("could not create cluster")

// IsCouldNotCreateClusterError asserts couldNotCreateClusterError.
func IsCouldNotCreateClusterError(err error) bool {
	return errgo.Cause(err) == couldNotCreateClusterError
}

// errors for cluster deletion

// couldNotDeleteClusterError should be used if the API call to delete a
// cluster has been responded with status >= 400
var couldNotDeleteClusterError = errgo.New("could not delete cluster")

// IsCouldNotDeleteClusterError asserts couldNotDeleteClusterError.
func IsCouldNotDeleteClusterError(err error) bool {
	return errgo.Cause(err) == couldNotDeleteClusterError
}

// Errors for scaling a cluster

// couldNotScaleClusterError should be used if the API call to scale a cluster
// has been responded with status >= 400
var couldNotScaleClusterError = errgo.New("could not scale cluster")

// IsCouldNotScaleClusterError asserts couldNotScaleClusterError.
func IsCouldNotScaleClusterError(err error) bool {
	return errgo.Cause(err) == couldNotScaleClusterError
}

// cannotScaleBelowMinimumWorkersError means the user tries to scale to less
// nodes than allowed
var cannotScaleBelowMinimumWorkersError = errgo.New("cannot scale below minimum amount of workers")

// IsCannotScaleBelowMinimumWorkersError asserts cannotScaleBelowMinimumWorkersError.
func IsCannotScaleBelowMinimumWorkersError(err error) bool {
	return errgo.Cause(err) == cannotScaleBelowMinimumWorkersError
}

// user has mixed incompatible settings related to different providers
var incompatibleSettingsError = errgo.New("incompatible mix of settings used")

// IsIncompatibleSettingsError asserts incompatibleSettingsError.
func IsIncompatibleSettingsError(err error) bool {
	return errgo.Cause(err) == incompatibleSettingsError
}

// endpointMissingError means the user has not given an endpoint where expected
var endpointMissingError = errgo.New("no endpoint given")

// IsEndpointMissingError asserts endpointMissingError.
func IsEndpointMissingError(err error) bool {
	return errgo.Cause(err) == endpointMissingError
}

// emptyPasswordError means the password supplied by the user was empty
var emptyPasswordError = errgo.New("empty password given")

// IsEmptyPasswordError asserts emptyPasswordError.
func IsEmptyPasswordError(err error) bool {
	return errgo.Cause(err) == emptyPasswordError
}

// tokenArgumentNotApplicableError means the user used --auth-token argument
// but it wasn't permitted for that command
var tokenArgumentNotApplicableError = errgo.New("token argument cannot be used here")

// IsTokenArgumentNotApplicableError asserts tokenArgumentNotApplicableError.
func IsTokenArgumentNotApplicableError(err error) bool {
	return errgo.Cause(err) == tokenArgumentNotApplicableError
}

// noEmailArgumentGivenError means the email argument was required
// but not given/empty
var noEmailArgumentGivenError = errgo.New("no email argument given")

// IsNoEmailArgumentGivenError asserts noEmailArgumentGivenError
func IsNoEmailArgumentGivenError(err error) bool {
	return errgo.Cause(err) == noEmailArgumentGivenError
}

// invalidCredentialsError means the user's credentials could not be verified
// by the API
var invalidCredentialsError = errgo.New("invalid credentials submitted")

// IsInvalidCredentialsError asserts invalidCredentialsError
func IsInvalidCredentialsError(err error) bool {
	return errgo.Cause(err) == invalidCredentialsError
}

// kubectlMissingError means that the 'kubectl' executable is not available
var kubectlMissingError = errgo.New("kubectl not installed")

// IsKubectlMissingError asserts kubectlMissingError
func IsKubectlMissingError(err error) bool {
	return errgo.Cause(err) == kubectlMissingError
}

// invalidReleaseError should be issued if the user selects a release that is
// either not active or does not exist
var invalidReleaseError = errgo.New("invalid release")

// IsInvalidReleaseError asserts invalidReleaseError
func IsInvalidReleaseError(err error) bool {
	return errgo.Cause(err) == invalidReleaseError
}
