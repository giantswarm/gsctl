package clienterror

import "net/http"

// IsMalformedResponseError checks whether the error is
// "Malformed response", which can mean several things.
func IsMalformedResponseError(err error) bool {
	return err.Error() == "Malformed response"
}

// IsBadRequestError checks whether the error
// is an HTTP 400 error.
func IsBadRequestError(err *APIError) bool {
	return err.HTTPStatusCode == http.StatusBadRequest
}

// IsAccessForbiddenError checks whether the error
// is an HTTP 403 error.
func IsAccessForbiddenError(err *APIError) bool {
	return err.HTTPStatusCode == http.StatusForbidden
}

// IsNotFoundError checks whether the error
// is an HTTP 404 error.
func IsNotFoundError(err *APIError) bool {
	return err.HTTPStatusCode == http.StatusNotFound
}
