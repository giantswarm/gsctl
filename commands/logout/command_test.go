package logout

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/testutils"
)

// Test_LogoutValidToken tests the logout for a valid token
func Test_LogoutValidToken(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been successfully deleted."}`))
	}))
	defer mockServer.Close()

	logoutArgs := Arguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	err = logout(logoutArgs)
	if err != nil {
		t.Error(err)
	}
}

// Test_LogoutInvalidToken tests the logout for an invalid token
func Test_LogoutInvalidToken(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "PERMISSION_DENIED", "message": "Nope"}`))
	}))
	defer mockServer.Close()

	logoutArgs := Arguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	err = logout(logoutArgs)

	if !clienterror.IsUnauthorizedError(err) {
		t.Errorf("Unexpected error type. Expected IsUnauthorizedError, got: %#v", err)
	}
}

// Test_LogoutSSO tests the logout for an user authenticated with SSO
func Test_LogoutSSO(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("This shouldn't have happened!")
	}))
	defer mockServer.Close()

	var configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    alias: frog
    provider: aws
    refresh_token: 8PyriiiTwCunDCK4tB7E4STRVuN-A3Rg7NaVIrZHmWWD5
    auth_scheme: Bearer
    token: eyJhbGcasdi1huojbnkjasd123jkbdkajs
updated: 2017-09-29T11:23:15+02:00
selected_endpoint: ` + mockServer.URL

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Error(err)
	}

	logoutArgs := Arguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
		scheme:      "Bearer",
	}

	err = logout(logoutArgs)
	if err != nil {
		t.Error(err)
	}
}

// Test_LogoutCommand simply calls the functions cobra would call,
// with a temporary config path and mock server as endpoint.
func Test_LogoutCommand(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been successfully deleted."}`))
	}))
	defer mockServer.Close()

	// config
	configYAML := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Error(err)
	}

	err = Command.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
}
