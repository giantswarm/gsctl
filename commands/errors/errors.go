// Package errors defines errors that can occur in command/input validation and
// execution and ways to handle those.
package errors

import (
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/microerror"
)

// UnknownError should be thrown if we have no idea what went wrong.
var UnknownError = &microerror.Error{
	Kind: "UnknownError",
}

// IsUnknownError asserts UnknownError.
func IsUnknownError(err error) bool {
	return microerror.Cause(err) == UnknownError
}

// CouldNotCreateClientError means that a client could not be created
var CouldNotCreateClientError = &microerror.Error{
	Kind: "CouldNotCreateClientError",
}

// IsCouldNotCreateClientError asserts CouldNotCreateClientError.
func IsCouldNotCreateClientError(err error) bool {
	return microerror.Cause(err) == CouldNotCreateClientError
}

// NotLoggedInError means that the user is currently not authenticated
var NotLoggedInError = &microerror.Error{
	Kind: "NotLoggedInError",
}

// IsNotLoggedInError asserts NotLoggedInError.
func IsNotLoggedInError(err error) bool {
	return microerror.Cause(err) == NotLoggedInError
}

// UserAccountInactiveError means that the user account is marked as inative by the API
var UserAccountInactiveError = &microerror.Error{
	Kind: "UserAccountInactiveError",
}

// IsUserAccountInactiveError asserts UserAccountInactiveError.
func IsUserAccountInactiveError(err error) bool {
	return microerror.Cause(err) == UserAccountInactiveError
}

// CommandAbortedError means that the user has aborted a command or input
var CommandAbortedError = &microerror.Error{
	Kind: "CommandAbortedError",
}

// IsCommandAbortedError asserts CommandAbortedError
func IsCommandAbortedError(err error) bool {
	return microerror.Cause(err) == CommandAbortedError
}

// ConflictingFlagsError means that the user combined command line options
// that are incompatible
var ConflictingFlagsError = &microerror.Error{
	Desc: "Some of the command line flags used cannot be combined.",
	Kind: "ConflictingFlagsError",
}

// IsConflictingFlagsError asserts ConflictingFlagsError.
func IsConflictingFlagsError(err error) bool {
	return microerror.Cause(err) == ConflictingFlagsError
}

// DesiredEqualsCurrentStateError means that the user described a desired
// state which is equal to the current state.
var DesiredEqualsCurrentStateError = &microerror.Error{
	Kind: "DesiredEqualsCurrentStateError",
}

// IsDesiredEqualsCurrentStateError asserts DesiredEqualsCurrentStateError.
func IsDesiredEqualsCurrentStateError(err error) bool {
	return microerror.Cause(err) == DesiredEqualsCurrentStateError
}

// ClusterIDMissingError means a required cluster ID has not been given as input
var ClusterIDMissingError = &microerror.Error{
	Kind: "ClusterIDMissingError",
}

// IsClusterIDMissingError asserts ClusterIDMissingError.
func IsClusterIDMissingError(err error) bool {
	return microerror.Cause(err) == ClusterIDMissingError
}

// NodePoolIDMissingError means a required node pool ID has not been given as input
var NodePoolIDMissingError = &microerror.Error{
	Kind: "NodePoolIDMissingError",
}

// IsNodePoolIDMissingError asserts NodePoolIDMissingError.
func IsNodePoolIDMissingError(err error) bool {
	return microerror.Cause(err) == NodePoolIDMissingError
}

// NodePoolIDMalformedError means a cluster/nodepool ID tuple has been formatted the wrong way.
var NodePoolIDMalformedError = &microerror.Error{
	Kind: "NodePoolIDMalformedError",
}

// IsNodePoolIDMalformedError asserts IsNodePoolIDMalformedError.
func IsNodePoolIDMalformedError(err error) bool {
	return microerror.Cause(err) == NodePoolIDMalformedError
}

// ClusterNotFoundError means that a given cluster does not exist
var ClusterNotFoundError = &microerror.Error{
	Kind: "ClusterNotFoundError",
}

// IsClusterNotFoundError asserts ClusterNotFoundError.
func IsClusterNotFoundError(err error) bool {
	c := microerror.Cause(err)
	if clienterror.IsNotFoundError(err) {
		return true
	}
	if c == ClusterNotFoundError {
		return true
	}

	return false
}

// NodePoolNotFoundError means that a node pool the user wants to interact with does not exist.
var NodePoolNotFoundError = &microerror.Error{
	Kind: "NodePoolNotFoundError",
}

// IsNodePoolNotFound asserts NodePoolNotFoundError.
func IsNodePoolNotFound(err error) bool {
	return microerror.Cause(err) == NodePoolNotFoundError
}

// ReleaseVersionMissingError means the required release version argument is missing
var ReleaseVersionMissingError = &microerror.Error{
	Kind: "ReleaseVersionMissingError",
}

// IsReleaseVersionMissingError asserts ReleaseVersionMissingError.
func IsReleaseVersionMissingError(err error) bool {
	return microerror.Cause(err) == ReleaseVersionMissingError
}

// ReleaseNotFoundError means that a given release does not exist.
var ReleaseNotFoundError = &microerror.Error{
	Kind: "ReleaseNotFoundError",
}

// IsReleaseNotFoundError asserts ReleaseNotFoundError.
func IsReleaseNotFoundError(err error) bool {
	return microerror.Cause(err) == ReleaseNotFoundError
}

// InternalServerError should only be used in case of server communication
// being responded to with a response status >= 500.
// See also: unknownError
var InternalServerError = &microerror.Error{
	Kind: "InternalServerError",
}

// IsInternalServerError asserts InternalServerError.
func IsInternalServerError(err error) bool {
	c := microerror.Cause(err)
	if clienterror.IsInternalServerError(err) {
		return true
	}
	if c == InternalServerError {
		return true
	}

	return false
}

// NoResponseError means the server side has not returned a response.
var NoResponseError = &microerror.Error{
	Kind: "NoResponseError",
}

// IsNoResponseError asserts NoResponseError.
func IsNoResponseError(err error) bool {
	return microerror.Cause(err) == NoResponseError
}

// NotAuthorizedError means that an API action could not be performed due to
// an authorization problem (usually a HTTP 401 error)
var NotAuthorizedError = &microerror.Error{
	Kind: "NotAuthorizedError",
}

// IsNotAuthorizedError asserts NotAuthorizedError.
func IsNotAuthorizedError(err error) bool {
	c := microerror.Cause(err)
	if clienterror.IsUnauthorizedError(err) {
		return true
	}
	if c == NotAuthorizedError {
		return true
	}

	return false
}

// Errors for cluster creation

// NotEnoughWorkerNodesError means that the user has specified a too low
// number of worker nodes for a cluster
var NotEnoughWorkerNodesError = &microerror.Error{
	Kind: "NotEnoughWorkerNodesError",
}

// IsNotEnoughWorkerNodesError asserts NotEnoughWorkerNodesError.
func IsNotEnoughWorkerNodesError(err error) bool {
	return microerror.Cause(err) == NotEnoughWorkerNodesError
}

// ClusterOwnerMissingError means that the user has not specified an owner organization
// for a new cluster
var ClusterOwnerMissingError = &microerror.Error{
	Kind: "ClusterOwnerMissingError",
}

// IsClusterOwnerMissingError asserts ClusterOwnerMissingError.
func IsClusterOwnerMissingError(err error) bool {
	return microerror.Cause(err) == ClusterOwnerMissingError
}

// OrganizationNotFoundError means that the specified organization could not be found
var OrganizationNotFoundError = &microerror.Error{
	Kind: "OrganizationNotFoundError",
}

// IsOrganizationNotFoundError asserts OrganizationNotFoundError
func IsOrganizationNotFoundError(err error) bool {
	c := microerror.Cause(err)
	if clienterror.IsNotFoundError(err) {
		return true
	}
	if c == OrganizationNotFoundError {
		return true
	}

	return false
}

// OrganizationNotSpecifiedError means that the user has not specified an organization to work with
var OrganizationNotSpecifiedError = &microerror.Error{
	Kind: "OrganizationNotSpecifiedError",
}

// IsOrganizationNotSpecifiedError asserts OrganizationNotSpecifiedError
func IsOrganizationNotSpecifiedError(err error) bool {
	return microerror.Cause(err) == OrganizationNotSpecifiedError
}

// CredentialNotFoundError means that the specified credential could not be found
var CredentialNotFoundError = &microerror.Error{
	Kind: "CredentialNotFoundError",
}

// IsCredentialNotFoundError asserts CredentialNotFoundError
func IsCredentialNotFoundError(err error) bool {
	return microerror.Cause(err) == CredentialNotFoundError
}

// YAMLFileNotReadableError means a YAML file was not readable
var YAMLFileNotReadableError = &microerror.Error{
	Kind: "YAMLFileNotReadableError",
}

// IsYAMLFileNotReadable asserts YAMLFileNotReadableError.
func IsYAMLFileNotReadable(err error) bool {
	return microerror.Cause(err) == YAMLFileNotReadableError
}

// YAMLNotParseableError means a YAML file was not readable
var YAMLNotParseableError = &microerror.Error{
	Kind: "YAMLNotParseableError",
}

// IsYAMLNotParseable asserts YAMLNotParseableError.
func IsYAMLNotParseable(err error) bool {
	return microerror.Cause(err) == YAMLNotParseableError
}

// CouldNotCreateJSONRequestBodyError occurs when we could not create a JSON
// request body based on the input we have, so something in out input attributes
// is wrong.
var CouldNotCreateJSONRequestBodyError = &microerror.Error{
	Kind: "CouldNotCreateJSONRequestBodyError",
}

// IsCouldNotCreateJSONRequestBodyError asserts CouldNotCreateJSONRequestBodyError.
func IsCouldNotCreateJSONRequestBodyError(err error) bool {
	return microerror.Cause(err) == CouldNotCreateJSONRequestBodyError
}

// CouldNotCreateClusterError should be used if the API call to create a
// cluster has been responded with status >= 400 and none of the other
// more specific errors apply.
var CouldNotCreateClusterError = &microerror.Error{
	Kind: "CouldNotCreateClusterError",
}

// IsCouldNotCreateClusterError asserts CouldNotCreateClusterError.
func IsCouldNotCreateClusterError(err error) bool {
	return microerror.Cause(err) == CouldNotCreateClusterError
}

// BadRequestError should be used when the server returns status 400 on cluster creation.
var BadRequestError = &microerror.Error{
	Kind: "BadRequestError",
}

// IsBadRequestError asserts BadRequestError
func IsBadRequestError(err error) bool {
	return microerror.Cause(err) == BadRequestError
}

// errors for cluster deletion

// CouldNotDeleteClusterError should be used if the API call to delete a
// cluster has been responded with status >= 400
var CouldNotDeleteClusterError = &microerror.Error{
	Kind: "CouldNotDeleteClusterError",
}

// IsCouldNotDeleteClusterError asserts CouldNotDeleteClusterError.
func IsCouldNotDeleteClusterError(err error) bool {
	return microerror.Cause(err) == CouldNotDeleteClusterError
}

// Errors for scaling a cluster

// CouldNotScaleClusterError should be used if the API call to scale a cluster
// has been responded with status >= 400
var CouldNotScaleClusterError = &microerror.Error{
	Kind: "CouldNotScaleClusterError",
}

// IsCouldNotScaleClusterError asserts CouldNotScaleClusterError.
func IsCouldNotScaleClusterError(err error) bool {
	return microerror.Cause(err) == CouldNotScaleClusterError
}

// APIError is happening when an error occurs in the API
var APIError = &microerror.Error{
	Kind: "APIError",
}

// IsAPIError asserts APIError.
func IsAPIError(err error) bool {
	c := microerror.Cause(err)
	_, ok := c.(*clienterror.APIError)
	if ok {
		return true
	}
	if c == APIError {
		return true
	}

	return false
}

// CannotScaleBelowMinimumWorkersError means the user tries to scale to less
// nodes than allowed
var CannotScaleBelowMinimumWorkersError = &microerror.Error{
	Kind: "CannotScaleBelowMinimumWorkersError",
}

// IsCannotScaleBelowMinimumWorkersError asserts CannotScaleBelowMinimumWorkersError.
func IsCannotScaleBelowMinimumWorkersError(err error) bool {
	return microerror.Cause(err) == CannotScaleBelowMinimumWorkersError
}

// CannotScaleClusterError means the user tries to scale a cluster that does not support
// scaling, e. g. because it is a v5 cluster (node pools).
var CannotScaleClusterError = &microerror.Error{
	Kind: "CannotScaleClusterError",
}

// IsCannotScaleCluster asserts CannotScaleClusterError.
func IsCannotScaleCluster(err error) bool {
	return microerror.Cause(err) == CannotScaleClusterError
}

// IncompatibleSettingsError means user has mixed incompatible settings related to different providers.
var IncompatibleSettingsError = &microerror.Error{
	Kind: "IncompatibleSettingsError",
}

// IsIncompatibleSettings asserts IncompatibleSettingsError.
func IsIncompatibleSettings(err error) bool {
	return microerror.Cause(err) == IncompatibleSettingsError
}

// EndpointMissingError means the user has not given an endpoint where expected
var EndpointMissingError = &microerror.Error{
	Kind: "EndpointMissingError",
}

// IsEndpointMissingError asserts EndpointMissingError.
func IsEndpointMissingError(err error) bool {
	return microerror.Cause(err) == EndpointMissingError
}

// EmptyPasswordError means the password supplied by the user was empty
var EmptyPasswordError = &microerror.Error{
	Kind: "EmptyPasswordError",
}

// IsEmptyPasswordError asserts EmptyPasswordError.
func IsEmptyPasswordError(err error) bool {
	return microerror.Cause(err) == EmptyPasswordError
}

// TokenArgumentNotApplicableError means the user used --auth-token argument
// but it wasn't permitted for that command
var TokenArgumentNotApplicableError = &microerror.Error{
	Kind: "TokenArgumentNotApplicableError",
}

// IsTokenArgumentNotApplicableError asserts TokenArgumentNotApplicableError.
func IsTokenArgumentNotApplicableError(err error) bool {
	return microerror.Cause(err) == TokenArgumentNotApplicableError
}

// PasswordArgumentNotApplicableError means the user used --password argument
// but it wasn't permitted for that command
var PasswordArgumentNotApplicableError = &microerror.Error{
	Kind: "PasswordArgumentNotApplicableError",
}

// IsPasswordArgumentNotApplicableError asserts PasswordArgumentNotApplicableError.
func IsPasswordArgumentNotApplicableError(err error) bool {
	return microerror.Cause(err) == PasswordArgumentNotApplicableError
}

// NoEmailArgumentGivenError means the email argument was required
// but not given/empty
var NoEmailArgumentGivenError = &microerror.Error{
	Kind: "NoEmailArgumentGivenError",
}

// IsNoEmailArgumentGivenError asserts NoEmailArgumentGivenError
func IsNoEmailArgumentGivenError(err error) bool {
	return microerror.Cause(err) == NoEmailArgumentGivenError
}

// AccessForbiddenError means the client has been denied access to the API endpoint
// with a HTTP 403 error
var AccessForbiddenError = &microerror.Error{
	Kind: "AccessForbiddenError",
}

// IsAccessForbiddenError asserts AccessForbiddenError
func IsAccessForbiddenError(err error) bool {
	c := microerror.Cause(err)
	if clienterror.IsAccessForbiddenError(err) {
		return true
	}
	if c == AccessForbiddenError {
		return true
	}

	return false
}

// InvalidCredentialsError means the user's credentials could not be verified
// by the API
var InvalidCredentialsError = &microerror.Error{
	Kind: "InvalidCredentialsError",
}

// IsInvalidCredentialsError asserts InvalidCredentialsError
func IsInvalidCredentialsError(err error) bool {
	return microerror.Cause(err) == InvalidCredentialsError
}

// KubectlMissingError means that the 'kubectl' executable is not available
var KubectlMissingError = &microerror.Error{
	Kind: "KubectlMissingError",
}

// IsKubectlMissingError asserts KubectlMissingError
func IsKubectlMissingError(err error) bool {
	return microerror.Cause(err) == KubectlMissingError
}

// CouldNotWriteFileError is used when an attempt to write some file fails
var CouldNotWriteFileError = &microerror.Error{
	Kind: "CouldNotWriteFileError",
}

// IsCouldNotWriteFileError asserts CouldNotWriteFileError
func IsCouldNotWriteFileError(err error) bool {
	return microerror.Cause(err) == CouldNotWriteFileError
}

// UnspecifiedAPIError means an API error has occurred which we can't or don't
// need to categorize any further.
var UnspecifiedAPIError = &microerror.Error{
	Kind: "UnspecifiedAPIError",
}

// IsUnspecifiedAPIError asserts UnspecifiedAPIError
func IsUnspecifiedAPIError(err error) bool {
	return microerror.Cause(err) == UnspecifiedAPIError
}

// NoUpgradeAvailableError means that the user wanted to start an upgrade, but
// there is no newer version available for the given cluster
var NoUpgradeAvailableError = &microerror.Error{
	Kind: "NoUpgradeAvailableError",
	Desc: "no upgrade available for the current version",
}

// IsNoUpgradeAvailableError asserts NoUpgradeAvailableError
func IsNoUpgradeAvailableError(err error) bool {
	return microerror.Cause(err) == NoUpgradeAvailableError
}

// CouldNotUpgradeClusterError is thrown when a cluster upgrade failed.
var CouldNotUpgradeClusterError = &microerror.Error{
	Kind: "CouldNotUpgradeClusterError",
	Desc: "could not upgrade cluster",
}

// IsCouldNotUpgradeClusterError asserts CouldNotUpgradeClusterError
func IsCouldNotUpgradeClusterError(err error) bool {
	return microerror.Cause(err) == CouldNotUpgradeClusterError
}

// InvalidCNPrefixError means the user has used bad characters in the CN prefix argument
var InvalidCNPrefixError = &microerror.Error{
	Kind: "InvalidCNPrefixError",
}

// IsInvalidCNPrefixError asserts InvalidCNPrefixError
func IsInvalidCNPrefixError(err error) bool {
	return microerror.Cause(err) == InvalidCNPrefixError
}

// InvalidDurationError means that a user-provided duration string could not be parsed
var InvalidDurationError = &microerror.Error{
	Kind: "InvalidDurationError",
}

// IsInvalidDurationError asserts InvalidDurationError
func IsInvalidDurationError(err error) bool {
	return microerror.Cause(err) == InvalidDurationError
}

// DurationExceededError is thrown when a duration value is larger than can be represented internally
var DurationExceededError = &microerror.Error{
	Kind: "DurationExceededError",
}

// IsDurationExceededError asserts DurationExceededError
func IsDurationExceededError(err error) bool {
	return microerror.Cause(err) == DurationExceededError
}

// SSOError means something went wrong during the SSO process
var SSOError = &microerror.Error{
	Kind: "SSOError",
}

// IsSSOError asserts SSOError
func IsSSOError(err error) bool {
	return microerror.Cause(err) == SSOError
}

// ProviderNotSupportedError means that the intended action is not possible with
// the installation's provider.
var ProviderNotSupportedError = &microerror.Error{
	Kind: "ProviderNotSupportedError",
}

// IsProviderNotSupportedError asserts ProviderNotSupportedError.
func IsProviderNotSupportedError(err error) bool {
	return microerror.Cause(err) == ProviderNotSupportedError
}

// RequiredFlagMissingError means that a required flag has not been set by the user.
var RequiredFlagMissingError = &microerror.Error{
	Kind: "RequiredFlagMissingError",
}

// IsRequiredFlagMissingError asserts RequiredFlagMissingError.
func IsRequiredFlagMissingError(err error) bool {
	return microerror.Cause(err) == RequiredFlagMissingError
}

// CredentialsAlreadySetError means the user tried setting credential to an org
// that has credentials already.
var CredentialsAlreadySetError = &microerror.Error{
	Kind: "CredentialsAlreadySetError",
}

// IsCredentialsAlreadySetError asserts CredentialsAlreadySetError.
func IsCredentialsAlreadySetError(err error) bool {
	return microerror.Cause(err) == CredentialsAlreadySetError
}

// UpdateCheckFailed means that checking for a newer gsctl version failed.
var UpdateCheckFailed = &microerror.Error{
	Kind: "UpdateCheckFailed",
}

// IsUpdateCheckFailed asserts UpdateCheckFailed.
func IsUpdateCheckFailed(err error) bool {
	return microerror.Cause(err) == UpdateCheckFailed
}

// ConflictingWorkerFlagsUsedError is raised when the deprecated --num-workers
// flag is used together with the new node count flags --workers-min and
// --workers-max.
var ConflictingWorkerFlagsUsedError = &microerror.Error{
	Kind: "ConflictingWorkerFlagsUsedError",
}

// IsConflictingWorkerFlagsUsed asserts ConflictingWorkerFlagsUsedError.
func IsConflictingWorkerFlagsUsed(err error) bool {
	return microerror.Cause(err) == ConflictingWorkerFlagsUsedError
}

// WorkersMinMaxInvalidError is raised when the value of the node count flag
// --workers-min is higher than the value of the node count flag --workers-max.
var WorkersMinMaxInvalidError = &microerror.Error{
	Kind: "WorkersMinMaxInvalidError",
	Desc: "min must not be higher than max",
}

// IsWorkersMinMaxInvalid asserts WorkersMinMaxInvalidError.
func IsWorkersMinMaxInvalid(err error) bool {
	return microerror.Cause(err) == WorkersMinMaxInvalidError
}

// OutputFormatInvalidError is raised when the user specifies an unsupported output format.
var OutputFormatInvalidError = &microerror.Error{
	Kind: "OutputFormatInvalidError",
}

// IsOutputFormatInvalid asserts OutputFormatInvalidError.
func IsOutputFormatInvalid(err error) bool {
	return microerror.Cause(err) == OutputFormatInvalidError
}

// ClusterDoesNotSupportNodePoolsError is raised when the user tries to do something with node pools
// on a cluster that does not support node pools.
var ClusterDoesNotSupportNodePoolsError = &microerror.Error{
	Kind: "ClusterDoesNotSupportNodePoolsError",
}

// IsClusterDoesNotSupportNodePools asserts ClusterDoesNotSupportNodePoolsError.
func IsClusterDoesNotSupportNodePools(err error) bool {
	return microerror.Cause(err) == ClusterDoesNotSupportNodePoolsError
}

// NoOpError is raised when the user calls a command without any meaningful
// parameters, resulting in no change/nothing done.
var NoOpError = &microerror.Error{
	Kind: "NoOpError",
	Desc: "Nothing to be done",
}

// IsNoOpError asserts NoOpError.
func IsNoOpError(err error) bool {
	return microerror.Cause(err) == NoOpError
}

// InvalidNodePoolIDArgumentError should be raised when the user gives a "clusterID/nodepoolID"
// argument that is syntactically incorrect.
var InvalidNodePoolIDArgumentError = &microerror.Error{
	Kind: "InvalidNodePoolIDArgumentError",
}

// IsInvalidNodePoolIDArgument asserts InvalidNodePoolIDArgumentError.
func IsInvalidNodePoolIDArgument(err error) bool {
	return microerror.Cause(err) == InvalidNodePoolIDArgumentError
}
