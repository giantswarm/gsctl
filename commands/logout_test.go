package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/viper"
)

// Test_LogoutValidToken tests the logout for a valid token
func Test_LogoutValidToken(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code": 10007, "status_text": "Resource deleted"}`))
	}))
	defer mockServer.Close()

	logoutArgs := logoutArguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	err := logout(logoutArgs)
	if err != nil {
		t.Error(err)
	}
}

// Test_LogoutInvalidToken tests the logout for an invalid token
func Test_LogoutInvalidToken(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{}`))
	}))
	defer mockServer.Close()

	logoutArgs := logoutArguments{
		apiEndpoint: mockServer.URL,
		token:       "test-token",
	}

	err := logout(logoutArgs)
	if err != nil {
		t.Error(err)
	}
}

// Test_LogoutCommand simply calls the functions cobra would call,
// with a temporary config path and mock server as endpoint.
func Test_LogoutCommand(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code": 10007, "status_text": "Resource deleted"}`))
	}))
	defer mockServer.Close()

	cmdAPIEndpoint = mockServer.URL
	cmdToken = "some-token"
	logoutValidationOutput(LogoutCommand, []string{})
	logoutOutput(LogoutCommand, []string{})
}
