package commands

import "github.com/giantswarm/microerror"

// Common errors and matcher functions for the "commands" package.

// unknownError should only be used if there is really no way to
// specify the error any further. Note that there is a more specific
// internalServerError.
var unknownError = &microerror.Error{
	Kind: "unknownError",
}

// IsUnknownError asserts unknownError.
func IsUnknownError(err error) bool {
	return microerror.Cause(err) == unknownError
}

// couldNotCreateClientError means that a client could not be created
var couldNotCreateClientError = &microerror.Error{
	Kind: "couldNotCreateClientError",
}

// IsCouldNotCreateClientError asserts couldNotCreateClientError.
func IsCouldNotCreateClientError(err error) bool {
	return microerror.Cause(err) == couldNotCreateClientError
}

// notLoggedInError means that the user is currently not authenticated
var notLoggedInError = &microerror.Error{
	Kind: "notLoggedInError",
}

// IsNotLoggedInError asserts notLoggedInError.
func IsNotLoggedInError(err error) bool {
	return microerror.Cause(err) == notLoggedInError
}

// userAccountInactiveError means that the user account is marked as inative by the API
var userAccountInactiveError = &microerror.Error{
	Kind: "userAccountInactiveError",
}

// IsUserAccountInactiveError asserts userAccountInactiveError.
func IsUserAccountInactiveError(err error) bool {
	return microerror.Cause(err) == userAccountInactiveError
}

// commandAbortedError means that the user has aborted a command or input
var commandAbortedError = &microerror.Error{
	Kind: "commandAbortedError",
}

// IsCommandAbortedError asserts commandAbortedError
func IsCommandAbortedError(err error) bool {
	return microerror.Cause(err) == commandAbortedError
}

// conflictingFlagsError means that the user combined command line options
// that are incompatible
var conflictingFlagsError = &microerror.Error{
	Desc: "Some of the command line flags used cannot be combined.",
	Kind: "conflictingFlagsError",
}

// IsConflictingFlagsError asserts conflictingFlagsError.
func IsConflictingFlagsError(err error) bool {
	return microerror.Cause(err) == conflictingFlagsError
}

// desiredEqualsCurrentStateError means that the user described a desired
// state which is equal to the current state.
var desiredEqualsCurrentStateError = &microerror.Error{
	Kind: "desiredEqualsCurrentStateError",
}

// IsDesiredEqualsCurrentStateError asserts desiredEqualsCurrentStateError.
func IsDesiredEqualsCurrentStateError(err error) bool {
	return microerror.Cause(err) == desiredEqualsCurrentStateError
}

// clusterIDMissingError means a required cluster ID has not been given as input
var clusterIDMissingError = &microerror.Error{
	Kind: "clusterIDMissingError",
}

// IsClusterIDMissingError asserts clusterIDMissingError.
func IsClusterIDMissingError(err error) bool {
	return microerror.Cause(err) == clusterIDMissingError
}

// clusterNotFoundError means that a given cluster does not exist
var clusterNotFoundError = &microerror.Error{
	Kind: "clusterNotFoundError",
}

// IsClusterNotFoundError asserts clusterNotFoundError.
func IsClusterNotFoundError(err error) bool {
	return microerror.Cause(err) == clusterNotFoundError
}

// internalServerError should only be used in case of server communication
// being responded to with a response status >= 500.
// See also: unknownError
var internalServerError = &microerror.Error{
	Kind: "internalServerError",
}

// IsInternalServerError asserts internalServerError.
func IsInternalServerError(err error) bool {
	return microerror.Cause(err) == internalServerError
}

// server side has not returned a response
var noResponseError = &microerror.Error{
	Kind: "noResponseError",
}

// IsNoResponseError asserts noResponseError.
func IsNoResponseError(err error) bool {
	return microerror.Cause(err) == noResponseError
}

// notAuthorizedError means that an API action could not be performed due to
// an authorization problem (usually a HTTP 401 error)
var notAuthorizedError = &microerror.Error{
	Kind: "notAuthorizedError",
}

// IsNotAuthorizedError asserts notAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	return microerror.Cause(err) == notAuthorizedError
}

// Errors for cluster creation

// numWorkerNodesMissingError means that the user has not specified how many
// worker nodes a new cluster should have
var numWorkerNodesMissingError = &microerror.Error{
	Kind: "numWorkerNodesMissingError",
}

// IsNumWorkerNodesMissingError asserts numWorkerNodesMissingError.
func IsNumWorkerNodesMissingError(err error) bool {
	return microerror.Cause(err) == numWorkerNodesMissingError
}

// notEnoughWorkerNodesError means that the user has specified a too low
// number of worker nodes for a cluster
var notEnoughWorkerNodesError = &microerror.Error{
	Kind: "notEnoughWorkerNodesError",
}

// IsNotEnoughWorkerNodesError asserts notEnoughWorkerNodesError.
func IsNotEnoughWorkerNodesError(err error) bool {
	return microerror.Cause(err) == notEnoughWorkerNodesError
}

// notEnoughCPUCoresPerWorkerError means the user did not request enough CPUs
// for the worker nodes
var notEnoughCPUCoresPerWorkerError = &microerror.Error{
	Kind: "notEnoughCPUCoresPerWorkerError",
}

// IsNotEnoughCPUCoresPerWorkerError asserts notEnoughCPUCoresPerWorkerError.
func IsNotEnoughCPUCoresPerWorkerError(err error) bool {
	return microerror.Cause(err) == notEnoughCPUCoresPerWorkerError
}

// notEnoughMemoryPerWorkerError means the user did not request enough RAM
// for the worker nodes
var notEnoughMemoryPerWorkerError = &microerror.Error{
	Kind: "notEnoughMemoryPerWorkerError",
}

// IsNotEnoughMemoryPerWorkerError asserts notEnoughMemoryPerWorkerError.
func IsNotEnoughMemoryPerWorkerError(err error) bool {
	return microerror.Cause(err) == notEnoughMemoryPerWorkerError
}

// notEnoughStoragePerWorkerError means the user did not request enough disk space
// for the worker nodes
var notEnoughStoragePerWorkerError = &microerror.Error{
	Kind: "notEnoughStoragePerWorkerError",
}

// IsNotEnoughStoragePerWorkerError asserts notEnoughStoragePerWorkerError.
func IsNotEnoughStoragePerWorkerError(err error) bool {
	return microerror.Cause(err) == notEnoughStoragePerWorkerError
}

// clusterOwnerMissingError means that the user has not specified an owner organization
// for a new cluster
var clusterOwnerMissingError = &microerror.Error{
	Kind: "clusterOwnerMissingError",
}

// IsClusterOwnerMissingError asserts clusterOwnerMissingError.
func IsClusterOwnerMissingError(err error) bool {
	return microerror.Cause(err) == clusterOwnerMissingError
}

// organizationNotFoundError means that the specified organization could not be found
var organizationNotFoundError = &microerror.Error{
	Kind: "organizationNotFoundError",
}

// IsOrganizationNotFoundError asserts organizationNotFoundError
func IsOrganizationNotFoundError(err error) bool {
	return microerror.Cause(err) == organizationNotFoundError
}

// organizationNotSpecifiedError means that the user has not specified an organization to work with
var organizationNotSpecifiedError = &microerror.Error{
	Kind: "organizationNotSpecifiedError",
}

// IsOrganizationNotSpecifiedError asserts organizationNotSpecifiedError
func IsOrganizationNotSpecifiedError(err error) bool {
	return microerror.Cause(err) == organizationNotSpecifiedError
}

// yamlFileNotReadableError means a YAML file was not readable
var yamlFileNotReadableError = &microerror.Error{
	Kind: "yamlFileNotReadableError",
}

// IsYAMLFileNotReadableError asserts yamlFileNotReadableError.
func IsYAMLFileNotReadableError(err error) bool {
	return microerror.Cause(err) == yamlFileNotReadableError
}

// couldNotCreateJSONRequestBodyError occurs when we could not create a JSON
// request body based on the input we have, so something in out input attributes
// is wrong.
var couldNotCreateJSONRequestBodyError = &microerror.Error{
	Kind: "couldNotCreateJSONRequestBodyError",
}

// IsCouldNotCreateJSONRequestBodyError asserts couldNotCreateJSONRequestBodyError.
func IsCouldNotCreateJSONRequestBodyError(err error) bool {
	return microerror.Cause(err) == couldNotCreateJSONRequestBodyError
}

// couldNotCreateClusterError should be used if the API call to create a
// cluster has been responded with status >= 400 and none of the other
// more specific errors apply.
var couldNotCreateClusterError = &microerror.Error{
	Kind: "couldNotCreateClusterError",
}

// IsCouldNotCreateClusterError asserts couldNotCreateClusterError.
func IsCouldNotCreateClusterError(err error) bool {
	return microerror.Cause(err) == couldNotCreateClusterError
}

// badRequestError should be used when the server returns status 400 on cluster creation.
var badRequestError = &microerror.Error{
	Kind: "badRequestError",
}

// IsBadRequestError asserts badRequestError
func IsBadRequestError(err error) bool {
	return microerror.Cause(err) == badRequestError
}

// errors for cluster deletion

// couldNotDeleteClusterError should be used if the API call to delete a
// cluster has been responded with status >= 400
var couldNotDeleteClusterError = &microerror.Error{
	Kind: "couldNotDeleteClusterError",
}

// IsCouldNotDeleteClusterError asserts couldNotDeleteClusterError.
func IsCouldNotDeleteClusterError(err error) bool {
	return microerror.Cause(err) == couldNotDeleteClusterError
}

// Errors for scaling a cluster

// couldNotScaleClusterError should be used if the API call to scale a cluster
// has been responded with status >= 400
var couldNotScaleClusterError = &microerror.Error{
	Kind: "couldNotScaleClusterError",
}

// IsCouldNotScaleClusterError asserts couldNotScaleClusterError.
func IsCouldNotScaleClusterError(err error) bool {
	return microerror.Cause(err) == couldNotScaleClusterError
}

// cannotScaleBelowMinimumWorkersError means the user tries to scale to less
// nodes than allowed
var cannotScaleBelowMinimumWorkersError = &microerror.Error{
	Kind: "cannotScaleBelowMinimumWorkersError",
}

// IsCannotScaleBelowMinimumWorkersError asserts cannotScaleBelowMinimumWorkersError.
func IsCannotScaleBelowMinimumWorkersError(err error) bool {
	return microerror.Cause(err) == cannotScaleBelowMinimumWorkersError
}

// user has mixed incompatible settings related to different providers
var incompatibleSettingsError = &microerror.Error{
	Kind: "incompatibleSettingsError",
}

// IsIncompatibleSettingsError asserts incompatibleSettingsError.
func IsIncompatibleSettingsError(err error) bool {
	return microerror.Cause(err) == incompatibleSettingsError
}

// endpointMissingError means the user has not given an endpoint where expected
var endpointMissingError = &microerror.Error{
	Kind: "endpointMissingError",
}

// IsEndpointMissingError asserts endpointMissingError.
func IsEndpointMissingError(err error) bool {
	return microerror.Cause(err) == endpointMissingError
}

// emptyPasswordError means the password supplied by the user was empty
var emptyPasswordError = &microerror.Error{
	Kind: "emptyPasswordError",
}

// IsEmptyPasswordError asserts emptyPasswordError.
func IsEmptyPasswordError(err error) bool {
	return microerror.Cause(err) == emptyPasswordError
}

// tokenArgumentNotApplicableError means the user used --auth-token argument
// but it wasn't permitted for that command
var tokenArgumentNotApplicableError = &microerror.Error{
	Kind: "tokenArgumentNotApplicableError",
}

// IsTokenArgumentNotApplicableError asserts tokenArgumentNotApplicableError.
func IsTokenArgumentNotApplicableError(err error) bool {
	return microerror.Cause(err) == tokenArgumentNotApplicableError
}

// passwordArgumentNotApplicableError means the user used --password argument
// but it wasn't permitted for that command
var passwordArgumentNotApplicableError = &microerror.Error{
	Kind: "passwordArgumentNotApplicableError",
}

// IsPasswordArgumentNotApplicableError asserts passwordArgumentNotApplicableError.
func IsPasswordArgumentNotApplicableError(err error) bool {
	return microerror.Cause(err) == passwordArgumentNotApplicableError
}

// noEmailArgumentGivenError means the email argument was required
// but not given/empty
var noEmailArgumentGivenError = &microerror.Error{
	Kind: "noEmailArgumentGivenError",
}

// IsNoEmailArgumentGivenError asserts noEmailArgumentGivenError
func IsNoEmailArgumentGivenError(err error) bool {
	return microerror.Cause(err) == noEmailArgumentGivenError
}

// accessForbiddenError means the client has been denied access to the API endpoint
// with a HTTP 403 error
var accessForbiddenError = &microerror.Error{
	Kind: "accessForbiddenError",
}

// IsAccessForbiddenError asserts accessForbiddenError
func IsAccessForbiddenError(err error) bool {
	return microerror.Cause(err) == accessForbiddenError
}

// invalidCredentialsError means the user's credentials could not be verified
// by the API
var invalidCredentialsError = &microerror.Error{
	Kind: "invalidCredentialsError",
}

// IsInvalidCredentialsError asserts invalidCredentialsError
func IsInvalidCredentialsError(err error) bool {
	return microerror.Cause(err) == invalidCredentialsError
}

// kubectlMissingError means that the 'kubectl' executable is not available
var kubectlMissingError = &microerror.Error{
	Kind: "kubectlMissingError",
}

// IsKubectlMissingError asserts kubectlMissingError
func IsKubectlMissingError(err error) bool {
	return microerror.Cause(err) == kubectlMissingError
}

// couldNotWriteFileError is used when an attempt to write some file fails
var couldNotWriteFileError = &microerror.Error{
	Kind: "couldNotWriteFileError",
}

// IsCouldNotWriteFileError asserts couldNotWriteFileError
func IsCouldNotWriteFileError(err error) bool {
	return microerror.Cause(err) == couldNotWriteFileError
}

// unspecifiedAPIError means an API error has occurred which we can't or don't
// need to categorize any further.
var unspecifiedAPIError = &microerror.Error{
	Kind: "unspecifiedAPIError",
}

// IsUnspecifiedAPIError asserts unspecifiedAPIError
func IsUnspecifiedAPIError(err error) bool {
	return microerror.Cause(err) == unspecifiedAPIError
}

// noUpgradeAvailableError means that the user wanted to start an upgrade, but
// there is no newer version available for the given cluster
var noUpgradeAvailableError = errgo.New("no upgrade available")

// IsNoUpgradeAvailableError asserts noUpgradeAvailableError
func IsNoUpgradeAvailableError(err error) bool {
	return errgo.Cause(err) == noUpgradeAvailableError
}

var couldNotUpgradeClusterError = errgo.New("could not upgrade cluster")

// IsCouldNotUpgradeClusterError asserts couldNotUpgradeClusterError
func IsCouldNotUpgradeClusterError(err error) bool {
	return errgo.Cause(err) == couldNotUpgradeClusterError
}

// invalidDurationError means that a user-provided duration string could not be parsed
var invalidDurationError = &microerror.Error{
	Kind: "invalidDurationError",
}

// IsInvalidDurationError asserts invalidDurationError
func IsInvalidDurationError(err error) bool {
	return microerror.Cause(err) == invalidDurationError
}

// durationExceededError is thrown when a duration value is larger than can be represented internally
var durationExceededError = &microerror.Error{
	Kind: "durationExceededError",
}

// IsDurationExceededError asserts durationExceededError
func IsDurationExceededError(err error) bool {
	return microerror.Cause(err) == durationExceededError
}

// ssoError means something went wrong during the SSO process
var ssoError = &microerror.Error{
	Kind: "ssoError",
}

// IsSSOError asserts ssoError
func IsSSOError(err error) bool {
	return microerror.Cause(err) == ssoError
}

// providerNotSupportedError means that the intended action is not possible with
// the installation's provider.
var providerNotSupportedError = &microerror.Error{
	Kind: "providerNotSupportedError",
}

// IsProviderNotSupportedError asserts providerNotSupportedError.
func IsProviderNotSupportedError(err error) bool {
	return microerror.Cause(err) == providerNotSupportedError
}

// requiredFlagMissingError means that a required flag has not been set by the user.
var requiredFlagMissingError = &microerror.Error{
	Kind: "requiredFlagMissingError",
}

// IsRequiredFlagMissingError asserts requiredFlagMissingError.
func IsRequiredFlagMissingError(err error) bool {
	return microerror.Cause(err) == requiredFlagMissingError
}

// credentialsAlreadySetError means the user tried setting credential to an org
// that has credentials already.
var credentialsAlreadySetError = &microerror.Error{
	Kind: "credentialsAlreadySetError",
}

// IsCredentialsAlreadySetError asserts credentialsAlreadySetError.
func IsCredentialsAlreadySetError(err error) bool {
	return microerror.Cause(err) == credentialsAlreadySetError
}
