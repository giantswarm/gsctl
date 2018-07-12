package clienterror

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"

	"github.com/go-openapi/runtime"

	"github.com/giantswarm/gsclientgen/client/auth_tokens"
)

// APIError is our structure to carry all error information we care about
// from the api client to the CLI
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

	// URL is the URL called with the request
	URL string

	// HTTPMethod is the HTTP method used
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
	convertedError, ok := err.(*auth_tokens.DeleteAuthTokenUnauthorized)
	if ok {
		return &APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			OriginalError:  convertedError,
			ErrorMessage:   "Unauthorized",
			ErrorDetails:   "The token presented was likely no longer valid. No further action is required.",
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
	// is url.Error
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
	// Note: We'd love to do type assertion here, but there is no exported type to assert with.
	errorType := reflect.TypeOf(err)
	if errorType != nil && errorType.String() == "context.deadlineExceededError" {
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
		ErrorMessage:  "Unknown error",
	}

	ae.ErrorDetails = "An error has occurred for which we don't have specific handling in place.\n"
	ae.ErrorDetails += "Please report this error to support@giantswarm including the command you\n"
	ae.ErrorDetails += "tried executing and the context information (gsctl info). Details:\n\n"
	ae.ErrorDetails += fmt.Sprintf("%#v", err)

	return ae
}
