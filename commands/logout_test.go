package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client/clienterror"
)

// Test_LogoutValidToken tests the logout for a valid token
func Test_LogoutValidToken(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been succesfully deleted."}`))
	}))
	defer mockServer.Close()

	logoutArgs := logoutArguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err = logout(logoutArgs)
	if err != nil {
		t.Error(err)
	}
}

// Test_LogoutInvalidToken tests the logout for an invalid token
func Test_LogoutInvalidToken(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "PERMISSION_DENIED", "message": "Nope"}`))
	}))
	defer mockServer.Close()

	logoutArgs := logoutArguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err = logout(logoutArgs)

	clientAPIErr, clientAPIErrOK := microerror.Cause(err).(*clienterror.APIError)
	if !clientAPIErrOK {
		t.Error("Type assertion to *clienterror.APIError failed. Error in unexpected type.")
	} else if clientAPIErr.HTTPStatusCode != http.StatusUnauthorized {
		t.Errorf("Unexpected HTTP status code: %d", clientAPIErr.HTTPStatusCode)
	}
}

// Test_LogoutCommand simply calls the functions cobra would call,
// with a temporary config path and mock server as endpoint.
func Test_LogoutCommand(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been succesfully deleted."}`))
	}))
	defer mockServer.Close()

	cmdAPIEndpoint = mockServer.URL
	cmdToken = "some-token"
	initClient()

	logoutValidationOutput(LogoutCommand, []string{})
	logoutOutput(LogoutCommand, []string{})
}
