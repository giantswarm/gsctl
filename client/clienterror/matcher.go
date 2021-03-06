package clienterror

import (
	"crypto/x509"
	"net/http"

	"github.com/giantswarm/microerror"
)

// IsMalformedResponse checks whether the error is
// "Malformed response", which can mean several things.
func IsMalformedResponse(err error) bool {
	return err.Error() == "Malformed response"
}

// IsBadRequestError checks whether the error
// is an HTTP 400 error.
func IsBadRequestError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusBadRequest
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusBadRequest
	}
	return false
}

// IsUnauthorizedError checks whether the error
// is an HTTP 401 error.
func IsUnauthorizedError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusUnauthorized
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusUnauthorized
	}
	return false
}

// IsAccessForbiddenError checks whether the error
// is an HTTP 403 error.
func IsAccessForbiddenError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusForbidden
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusForbidden
	}
	return false
}

// IsNotFoundError checks whether the error
// is an HTTP 404 error.
func IsNotFoundError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusNotFound
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusNotFound
	}
	return false
}

// IsConflictError checks whether the error
// is an HTTP 409 error.
func IsConflictError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusConflict
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusConflict
	}
	return false
}

// IsInternalServerError checks whether the error
// is an HTTP 500 error.
func IsInternalServerError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusInternalServerError
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusInternalServerError
	}
	return false
}

// IsServiceUnavailableError checks whether the error
// is an HTTP 503 error.
func IsServiceUnavailableError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		return clientErr.HTTPStatusCode == http.StatusServiceUnavailable
	}
	if apiErr, apiErrOK := microerror.Cause(err).(*APIError); apiErrOK {
		return apiErr.HTTPStatusCode == http.StatusServiceUnavailable
	}
	return false
}

// IsCertificateSignedByUnknownAuthorityError checks whether the error represents
// a x509.UnknownAuthorityError
func IsCertificateSignedByUnknownAuthorityError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		if _, certErrorOK := clientErr.OriginalError.(x509.UnknownAuthorityError); certErrorOK {
			return true
		}
	}

	return false
}

// IsCertificateHostnameError checks whether the error represents
// a x509.HostnameError
func IsCertificateHostnameError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		if _, certErrorOK := clientErr.OriginalError.(x509.HostnameError); certErrorOK {
			return true
		}
	}

	return false
}

// IsCertificateInvalidError checks whether the error represents
// a x509.UnknownAuthorityError
func IsCertificateInvalidError(err error) bool {
	if clientErr, ok := err.(*APIError); ok {
		if _, certErrorOK := clientErr.OriginalError.(x509.CertificateInvalidError); certErrorOK {
			return true
		}
	}

	return false
}
