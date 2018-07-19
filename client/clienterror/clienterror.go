package clienterror

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/go-openapi/runtime"

	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/client/clusters"
	"github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/gsclientgen/client/key_pairs"
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
	deleteAuthTokenUnauthorizedError, ok := err.(*auth_tokens.DeleteAuthTokenUnauthorized)
	if ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  deleteAuthTokenUnauthorizedError,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "The token presented was likely no longer valid. No further action is required.",
		}
	}

	// create auth token
	createAuthTokenUnauthorizedError, ok := err.(*auth_tokens.CreateAuthTokenUnauthorized)
	if ok {
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
	createClusterUnauthorizedErr, ok := err.(*clusters.AddClusterUnauthorized)
	if ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  createClusterUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to create a cluster for this organization.",
		}
	}
	createClusterDefaultErr, ok := err.(*clusters.AddClusterDefault)
	if ok {
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

	// create key pair
	createKeyPairUnauthorizedErr, ok := err.(*key_pairs.AddKeyPairUnauthorized)
	if ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  createKeyPairUnauthorizedErr,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to create a key pair for this cluster.",
		}
	}

	// get info
	getInfoUnauthorizedErr, ok := err.(*info.GetInfoUnauthorized)
	if ok {
		return &APIError{
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "You don't have permission to get information on this installation.",
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  getInfoUnauthorizedErr,
		}
	}
	getInfoDefaultErr, ok := err.(*info.GetInfoDefault)
	if ok {
		return &APIError{
			ErrorMessage:   getInfoDefaultErr.Error(),
			HTTPStatusCode: getInfoDefaultErr.Code(),
			OriginalError:  getInfoDefaultErr,
		}
	}

	// HTTP level error cases
	runtimeAPIError, runtimeAPIErrorOK := err.(*runtime.APIError)
	if runtimeAPIErrorOK {
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
	urlError, urlErrorOK := err.(*url.Error)
	if urlErrorOK {
		ae := &APIError{
			OriginalError: urlError,
			URL:           urlError.URL,
			HTTPMethod:    urlError.Op,
		}

		// is net.OpError
		netOpError, netOpErrorOK := urlError.Err.(*net.OpError)
		if netOpErrorOK {
			ae.OriginalError = netOpError

			// is net.DNSError
			netDNSError, netDNSErrorOK := netOpError.Err.(*net.DNSError)
			if netDNSErrorOK {
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

	// Return unspecific error
	ae := &APIError{
		OriginalError: err,
		ErrorMessage:  "Unknown error: " + err.Error(),
	}

	ae.ErrorDetails = "An error has occurred for which we don't have specific handling in place.\n"
	ae.ErrorDetails += "Please report this error to support@giantswarm including the command you\n"
	ae.ErrorDetails += "tried executing and the context information (gsctl info). Details:\n\n"
	ae.ErrorDetails += fmt.Sprintf("%#v", err)

	return ae
}
