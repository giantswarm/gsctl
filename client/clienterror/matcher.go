package clienterror

import "net/http"

// IsMalformedResponseError checks whether the error is
// "Malformed response", which can mean several things.
func IsMalformedResponseError(err error) bool {
	return err.Error() == "Malformed response"
}

// IsBadRequestError checks whether the error
// is an HTTP 400 error.
func IsBadRequestError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusBadRequest
	}
	return false
}

// IsUnauthorizedError checks whether the error
// is an HTTP 401 error.
func IsUnauthorizedError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusUnauthorized
	}
	return false
}

// IsAccessForbiddenError checks whether the error
// is an HTTP 403 error.
func IsAccessForbiddenError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusForbidden
	}
	return false
}

// IsNotFoundError checks whether the error
// is an HTTP 404 error.
func IsNotFoundError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusNotFound
	}
	return false
}

// IsConflictError checks whether the error
// is an HTTP 409 error.
func IsConflictError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusConflict
	}
	return false
}

// IsInternalServerError checks whether the error
// is an HTTP 500 error.
func IsInternalServerError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusInternalServerError
	}
	return false
}
