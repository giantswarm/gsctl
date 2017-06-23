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

var couldNotCreateClientError = errgo.New("could not create client")

// IsCouldNotCreateClientError asserts couldNotCreateClientError.
func IsCouldNotCreateClientError(err error) bool {
	return errgo.Cause(err) == couldNotCreateClientError
}

var notLoggedInError = errgo.New("user not logged in")

// IsNotLoggedInError asserts notLoggedInError.
func IsNotLoggedInError(err error) bool {
	return errgo.Cause(err) == notLoggedInError
}

var conflictingFlagsError = errgo.New("conflicting flags used")

// IsConflictingFlagsError asserts conflictingFlagsError.
func IsConflictingFlagsError(err error) bool {
	return errgo.Cause(err) == conflictingFlagsError
}

var clusterIDMissingError = errgo.New("cluster ID not specified")

// IsClusterIDMissingError asserts clusterIDMissingError.
func IsClusterIDMissingError(err error) bool {
	return errgo.Cause(err) == clusterIDMissingError
}

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

var notAuthorizedError = errgo.New("not authorized")

// IsNotAuthorizedError asserts notAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return errgo.Cause(err) == notAuthorizedError
}

// Errors for cluster creation

var numWorkerNodesMissingError = errgo.New("number of workers not specified")

// IsNumWorkerNodesMissingError asserts numWorkerNodesMissingError.
func IsNumWorkerNodesMissingError(err error) bool {
	return errgo.Cause(err) == numWorkerNodesMissingError
}

var notEnoughWorkerNodesError = errgo.New("not enough workers specified")

// IsNotEnoughWorkerNodesError asserts notEnoughWorkerNodesError.
func IsNotEnoughWorkerNodesError(err error) bool {
	return errgo.Cause(err) == notEnoughWorkerNodesError
}

var notEnoughCPUCoresPerWorkerError = errgo.New("not enough CPU cores per worker specified")

// IsNotEnoughCPUCoresPerWorkerError asserts notEnoughCPUCoresPerWorkerError.
func IsNotEnoughCPUCoresPerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughCPUCoresPerWorkerError
}

var notEnoughMemoryPerWorkerError = errgo.New("not enough memory per worker specified")

// IsNotEnoughMemoryPerWorkerError asserts notEnoughMemoryPerWorkerError.
func IsNotEnoughMemoryPerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughMemoryPerWorkerError
}

var notEnoughStoragePerWorkerError = errgo.New("not enough storage per worker specified")

// IsNotEnoughStoragePerWorkerError asserts notEnoughStoragePerWorkerError.
func IsNotEnoughStoragePerWorkerError(err error) bool {
	return errgo.Cause(err) == notEnoughStoragePerWorkerError
}

var clusterOwnerMissingError = errgo.New("no cluster owner specified")

// IsClusterOwnerMissingError asserts clusterOwnerMissingError.
func IsClusterOwnerMissingError(err error) bool {
	return errgo.Cause(err) == clusterOwnerMissingError
}

var yamlFileNotReadableError = errgo.New("could not read YAML file")

// IsYAMLFileNotReadableError asserts yamlFileNotReadableError.
func IsYAMLFileNotReadableError(err error) bool {
	return errgo.Cause(err) == yamlFileNotReadableError
}

var couldNotCreateJSONRequestBodyError = errgo.New("could not create JSON request body")

// IsCouldNotCreateJSONRequestBodyError asserts couldNotCreateJSONRequestBodyError.
func IsCouldNotCreateJSONRequestBodyError(err error) bool {
	return errgo.Cause(err) == couldNotCreateJSONRequestBodyError
}

// should be used if the API call to create a cluster has been responded with
// status >= 400
var couldNotCreateClusterError = errgo.New("could not create cluster")

// IsCouldNotCreateClusterError asserts couldNotCreateClusterError.
func IsCouldNotCreateClusterError(err error) bool {
	return errgo.Cause(err) == couldNotCreateClusterError
}

// errors for cluster deletion

// should be used if the API call to create a cluster has been responded with
// status >= 400
var couldNotDeleteClusterError = errgo.New("could not delete cluster")

// IsCouldNotDeleteClusterError asserts couldNotDeleteClusterError.
func IsCouldNotDeleteClusterError(err error) bool {
	return errgo.Cause(err) == couldNotDeleteClusterError
}
