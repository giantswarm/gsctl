package login

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/testutils"
)

// regularInfoResponse is a JSON snippet we use in several test cases
var regularInfoResponse = []byte(`{
	"general": {
		"installation_name": "codename",
		"provider": "aws",
		"datacenter": "eu-central-foo"
	},
	"workers": {
		"count_per_cluster": {
			"max": 20,
			"default": 3
		},
		"instance_type": {
			"options": ["m3.large", "m4.xlarge"],
			"default": "m3.large"
		}
	}
}`)

// Test_LoginValidPassword simulates a login with a valid email/password combination
func Test_LoginValidPassword(t *testing.T) {
	// we start with an empty config
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	// this server will respond positively in any case
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.String() == "/v4/info/" {
			w.Write(regularInfoResponse)
		} else {
			w.Write([]byte(`{"auth_token": "some-test-session-token"}`))
		}
	}))
	defer mockServer.Close()

	args := Arguments{
		apiEndpoint: mockServer.URL,
		email:       "email@example.com",
		password:    "test password",
		verbose:     true,
	}

	result, err := login(args)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if result.email != args.email {
		t.Errorf("Expected %q, got %q", args.email, result.email)
	}
	if result.token != "some-test-session-token" {
		t.Errorf("Expected 'some-test-session-token', got %q", result.token)
	}
	if result.alias != "codename" {
		t.Errorf("Expected alias 'codename', got %q", result.alias)
	}
	if result.loggedOutBefore == true {
		t.Error("result.loggedOutBefore was true, expected false")
	}
	if result.numEndpointsBefore != 0 {
		t.Error("Expected result.numEndpointsBefore to be 0, got", result.numEndpointsBefore)
	}
	if result.numEndpointsAfter != 1 {
		t.Error("Expected result.numEndpointsAfter to be 1, got", result.numEndpointsAfter)
	}
}

// Test_LoginInvalidPassword simulates a login with a bad email/password combination
func Test_LoginInvalidPassword(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
      "code": "PERMISSION_DENIED",
      "message": "Some bad credentials error message"
    }`))
	}))
	defer mockServer.Close()

	args := Arguments{
		apiEndpoint: mockServer.URL,
		email:       "email@example.com",
		password:    "bad password",
	}

	_, err = login(args)

	if !clienterror.IsUnauthorizedError(err) {
		t.Errorf("Expected error 401, got %#v", err)
	}
}

// Test_LoginWhenUserLoggedInBefore simulates an okay login when the user was
// logged in before.
func Test_LoginWhenUserLoggedInBefore(t *testing.T) {
	// this server will respond positively in any case
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.String() == "/v4/info/" {
			w.Write(regularInfoResponse)
		} else {
			w.Write([]byte(`{"auth_token": "another-test-session-token"}`))
		}
	}))
	defer mockServer.Close()

	// config
	yamlText := `endpoints:
  "` + mockServer.URL + `":
    email: email@foo.com
    token: token
selected_endpoint: "` + mockServer.URL + `"
`
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
	if err != nil {
		fmt.Printf(yamlText)
		t.Error(err)
	}

	args := Arguments{
		apiEndpoint: mockServer.URL,
		email:       "email@example.com",
		password:    "test password",
	}

	result, loginErr := login(args)
	if loginErr != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if result.email != args.email {
		t.Errorf("Expected %q, got %q", args.email, result.email)
	}
	if config.Config.Email != args.email {
		t.Errorf("Expected config email to be %q, got %q", args.email, config.Config.Email)
	}
	if result.token != "another-test-session-token" {
		t.Errorf("Expected 'another-test-session-token', got %q", result.token)
	}
	if !result.loggedOutBefore {
		t.Error("result.loggedOutBefore was false, expected true")
	}
}

// Test_LoginInactiveAccount simulates a login with an inactive/expired account
func Test_LoginInactiveAccount(t *testing.T) {
	// we start with an empty config
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	// mock server responding with a 400 Bad request
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "ACCOUNT_EXPIRED", "message": "Lorem ipsum"}`))
	}))
	defer mockServer.Close()

	args := Arguments{
		apiEndpoint: mockServer.URL,
		email:       "developer@giantswarm.io",
		password:    "test password",
	}

	_, err = login(args)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	convertedError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Error type assertion to *clienterror.APIError failed")
	}

	if !clienterror.IsUnauthorizedError(err) {
		t.Errorf("Expected error 401, got %#v", convertedError.HTTPStatusCode)
	}

	origErr, ok := convertedError.OriginalError.(*auth_tokens.CreateAuthTokenUnauthorized)
	if !ok {
		t.Error("Error type assertion to *auth_tokens.CreateAuthTokenUnauthorized failed")
	}
	if origErr.Payload.Code != "ACCOUNT_EXPIRED" {
		t.Errorf("Expected 'ACCOUNT_EXPIRED', got %#v", origErr.Payload.Code)
	}
}
