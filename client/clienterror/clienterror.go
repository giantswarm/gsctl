package clienterror

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-openapi/runtime"

	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/client/clusters"
	"github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/gsclientgen/client/key_pairs"
	"github.com/giantswarm/gsclientgen/client/organizations"
	"github.com/giantswarm/gsclientgen/client/releases"
)

// APIError is our structure to carry all error information we care about
// from the api client to the CLI.
type APIError struct {
	// HTTPStatusCode holds the HTTP response status code, if there was any.
	HTTPStatusCode int

	// OriginalError is the original error object which should contain
	// type-specific details.
	OriginalError error

	// ErrorMessage is a short, user-friendly error message we generate for
	// presenting details to the end user.
	ErrorMessage string

	// ErrorDetails is a longer text we MAY set additionally to help the user
	// understand and maybe solve the problem.
	ErrorDetails string

	// URL is the URL called with the request.
	URL string

	// HTTPMethod is the HTTP method used.
	HTTPMethod string

	// IsTimeout will be true if our error was a timeout error.
	IsTimeout bool

	// IsTemporary will be true if we think that a retry will help.
	IsTemporary bool
}

// Error returns the error message and allows us to use our APIError
// as an error type.
func (ae APIError) Error() string {
	return ae.ErrorMessage
}

// New creates a new APIError based on all the incoming error details. One
// goal here is to let handlers deal with only one type of error.
func New(err error) *APIError {
	if err == nil {
		return nil
	}

	// We first handle the most specific cases, which differ between operations.
	// When adding support for more API operations to the client, add handling
	// of any new specific error types here.
	if deleteAuthTokenUnauthorizedError, ok := err.(*auth_tokens.DeleteAuthTokenUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  deleteAuthTokenUnauthorizedError,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "The token presented was likely no longer valid. No further action is required.",
		}
	}

	// create auth token
	if createAuthTokenUnauthorizedError, ok := err.(*auth_tokens.CreateAuthTokenUnauthorized); ok {
		ae := &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  createAuthTokenUnauthorizedError,
			ErrorMessage:   "Bad credentials",
			ErrorDetails:   "The email and password presented don't match any known user credentials. Please check and try again.",
		}

		if createAuthTokenUnauthorizedError.Payload.Code == "ACCOUNT_EXPIRED" {
			ae.ErrorMessage = "Account expired"
			ae.ErrorDetails = "Please contact the Giant Swarm support team to help you out."
		}

		return ae
	}

	// create cluster
	if createClusterUnauthorizedErr, ok := err.(*clusters.AddClusterUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  createClusterUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to create a cluster for this organization.",
		}
	}
	if createClusterDefaultErr, ok := err.(*clusters.AddClusterDefault); ok {
		ae := &APIError{
			HTTPStatusCode: createClusterDefaultErr.Code(),
			OriginalError:  createClusterDefaultErr,
			ErrorMessage:   createClusterDefaultErr.Error(),
		}
		if ae.HTTPStatusCode == http.StatusNotFound {
			ae.ErrorMessage = "Organization does not exist"
			ae.ErrorDetails = "The organization to own the cluster does not exist. Please check the name."
		} else if ae.HTTPStatusCode == http.StatusBadRequest {
			ae.ErrorMessage = "Invalid parameters"
			ae.ErrorDetails = "The cluster cannot be created. Some parameter(s) are considered invalid.\n"
			ae.ErrorDetails += "Details: " + createClusterDefaultErr.Payload.Message
		}
		return ae
	}

	// delete cluster
	if deleteClusterUnauthorizedErr, ok := err.(*clusters.DeleteClusterUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  deleteClusterUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to delete this cluster.",
		}
	}
	if deleteClusterDefaultErr, ok := err.(*clusters.DeleteClusterDefault); ok {
		return &APIError{
			HTTPStatusCode: deleteClusterDefaultErr.Code(),
			OriginalError:  deleteClusterDefaultErr,
			ErrorMessage:   deleteClusterDefaultErr.Error(),
		}
	}

	// get clusters
	if getClustersUnauthorizedErr, ok := err.(*clusters.GetClustersUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getClustersUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to list clusters for this organization.",
		}
	}
	if getClustersDefaultErr, ok := err.(*clusters.GetClustersDefault); ok {
		return &APIError{
			HTTPStatusCode: getClustersDefaultErr.Code(),
			OriginalError:  getClustersDefaultErr,
			ErrorMessage:   getClustersDefaultErr.Error(),
		}
	}

	// get cluster
	if getClusterUnauthorizedErr, ok := err.(*clusters.GetClusterUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getClusterUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to access this cluster's details.",
		}
	}
	if getClusterNotFoundErr, ok := err.(*clusters.GetClusterNotFound); ok {
		return &APIError{
			HTTPStatusCode: http.StatusNotFound,
			OriginalError:  getClusterNotFoundErr,
			ErrorMessage:   "Cluster not found",
			ErrorDetails:   "The cluster with the given ID does not exist.",
		}
	}
	if getClusterDefaultErr, ok := err.(*clusters.GetClusterDefault); ok {
		return &APIError{
			HTTPStatusCode: getClusterDefaultErr.Code(),
			OriginalError:  getClusterDefaultErr,
			ErrorMessage:   getClusterDefaultErr.Error(),
		}
	}

	// create key pair
	if createKeyPairUnauthorizedErr, ok := err.(*key_pairs.AddKeyPairUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  createKeyPairUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to create a key pair for this cluster.",
		}
	}

	// get key pairs
	if getKeyPairsUnauthorizedErr, ok := err.(*key_pairs.GetKeyPairsUnauthorized); ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getKeyPairsUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to list key pairs for this cluster.",
		}
	}
	if getKeyPairsDefaultErr, ok := err.(*key_pairs.GetKeyPairsDefault); ok {
		return &APIError{
			HTTPStatusCode: getKeyPairsDefaultErr.Code(),
			OriginalError:  getKeyPairsDefaultErr,
			ErrorMessage:   getKeyPairsDefaultErr.Error(),
		}
	}

	// get info
	if getInfoUnauthorizedErr, ok := err.(*info.GetInfoUnauthorized); ok {
		return &APIError{
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to get information on this installation.",
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getInfoUnauthorizedErr,
		}
	}
	if getInfoDefaultErr, ok := err.(*info.GetInfoDefault); ok {
		return &APIError{
			ErrorMessage:   getInfoDefaultErr.Error(),
			HTTPStatusCode: getInfoDefaultErr.Code(),
			OriginalError:  getInfoDefaultErr,
		}
	}

	// get releases
	if getReleasesUnauthorized, ok := err.(*releases.GetReleasesUnauthorized); ok {
		return &APIError{
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to list releases on this installation.",
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getReleasesUnauthorized,
		}
	}

	// get organizations
	if getOrganizationsUnauthorized, ok := err.(*organizations.GetOrganizationsUnauthorized); ok {
		return &APIError{
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to list organizations in this installation.",
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getOrganizationsUnauthorized,
		}
	}
	if getOrganizationsDefault, ok := err.(*organizations.GetOrganizationsDefault); ok {
		return &APIError{
			ErrorMessage:   getOrganizationsDefault.Error(),
			HTTPStatusCode: getOrganizationsDefault.Code(),
			OriginalError:  getOrganizationsDefault,
		}
	}

	// add credentials
	if addCredentialsConflict, ok := err.(*organizations.AddCredentialsConflict); ok {
		return &APIError{
			ErrorMessage:   "The organization has credentials already",
			ErrorDetails:   "Credentials are immutable and an organization can only have one credential set.",
			HTTPStatusCode: http.StatusConflict,
			OriginalError:  addCredentialsConflict,
		}
	}
	if addCredentialsUnauthorized, ok := err.(*organizations.AddCredentialsUnauthorized); ok {
		return &APIError{
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You do not have permission to set credentials for this organization.",
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  addCredentialsUnauthorized,
		}
	}
	if addCredentialsDefault, ok := err.(*organizations.AddCredentialsDefault); ok {
		return &APIError{
			ErrorMessage:   addCredentialsDefault.Error(),
			HTTPStatusCode: addCredentialsDefault.Code(),
			OriginalError:  addCredentialsDefault,
		}
	}

	// HTTP level error cases
	if runtimeAPIError, ok := err.(*runtime.APIError); ok {
		ae := &APIError{
			HTTPStatusCode: runtimeAPIError.Code,
			OriginalError:  runtimeAPIError,
			ErrorMessage:   fmt.Sprintf("HTTP Error %d", runtimeAPIError.Code),
		}

		// special messages for some codes
		switch runtimeAPIError.Code {

		case http.StatusForbidden:
			ae.ErrorMessage = "Forbidden"
			ae.ErrorDetails = "The client has been denied access to the API endpoint with an HTTP status of 403.\n"
			ae.ErrorDetails += "Please make sure that you are in the right network or VPN. Once that is verified,\n"
			ae.ErrorDetails += "check back with Giant Swarm support that your network is permitted access."

		case http.StatusUnauthorized:
			ae.ErrorMessage = "Not authorized"
			ae.ErrorDetails = "You are not authorized for this action.\n"
			ae.ErrorDetails += "Please check whether you are logged in with the given endpoint."

		case http.StatusInternalServerError:
			ae.ErrorMessage = "Backend error"
			ae.ErrorDetails = "The backend responded with an HTTP 500 code, indicating an internal error on Giant Swarm's side.\n"
			ae.ErrorDetails += "Original error message: " + runtimeAPIError.Error() + "\n"
			ae.ErrorDetails += "Please report this problem to the Giant Swarm support team."
		}

		return ae
	}

	// Errors on levels lower than HTTP
	// is url.Error.
	if urlError, ok := err.(*url.Error); ok {
		ae := &APIError{
			OriginalError: urlError,
			URL:           urlError.URL,
			HTTPMethod:    urlError.Op,
		}

		// is net.OpError
		if netOpError, netOpErrorOK := urlError.Err.(*net.OpError); netOpErrorOK {
			ae.OriginalError = netOpError

			// is net.DNSError
			if netDNSError, netDNSErrorOK := netOpError.Err.(*net.DNSError); netDNSErrorOK {
				ae.OriginalError = netDNSError
				ae.IsTemporary = netDNSError.IsTemporary
				ae.IsTimeout = netDNSError.IsTimeout
				ae.ErrorMessage = "DNS error"
				ae.ErrorDetails = fmt.Sprintf("The host name '%s' cannot be resolved.\n", netDNSError.Name)
				ae.ErrorDetails += "Please make sure the endpoint URL you are using is correct."

				return ae
			}

			// dial error
			if netOpError.Op == "dial" {
				ae.ErrorMessage = "No connection to host"
				ae.ErrorDetails = fmt.Sprintf("The host '%s' cannot be reached.\n", netOpError.Addr.String())
				ae.ErrorDetails += "Please make sure that you are in the right network or VPN."

				return ae
			}
		}

		return ae
	}

	// Timeout / context deadline exceeded
	if err == context.DeadlineExceeded {
		ae := &APIError{
			OriginalError: err,
			IsTimeout:     true,
			IsTemporary:   true,
			ErrorMessage:  "Request timed out",
			ErrorDetails:  "Something took longer than expected. Please try again.",
		}

		return ae
	}

	// Response parser error - likely indicating that we didn't talk to the API, but some
	// proxy instead which didn't respond with JSON but plain text.
	// Example: '(*models.V4GenericResponse) is not supported by the TextConsumer, can be resolved by supporting TextUnmarshaler interface'
	if strings.Contains(err.Error(), "TextConsumer") && strings.Contains(err.Error(), "TextUnmarshaler") {
		ae := &APIError{
			OriginalError: err,
			ErrorMessage:  "Malformed response",
		}

		ae.ErrorDetails = "The response we received did not match the expected format. The reason could be that we\n"
		ae.ErrorDetails += "don't have access to the actual API server. Please check whether your network has access\n"
		ae.ErrorDetails += "using the 'gsctl ping' command with the according endpoint.\n"

		return ae
	}

	// Return unspecific error
	ae := &APIError{
		OriginalError: err,
		ErrorMessage:  "Unknown error: " + err.Error(),
	}

	ae.ErrorDetails = "An error has occurred for which we don't have specific handling in place.\n"
	ae.ErrorDetails += "Please report this error to support@giantswarm.io including the command you\n"
	ae.ErrorDetails += "tried executing and the context information (gsctl info). Details:\n\n"
	ae.ErrorDetails += fmt.Sprintf("%#v", err)

	return ae
}
