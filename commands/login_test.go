package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// Test_LoginValidPassword simulates a login with a valid email/password combination
func Test_LoginValidPassword(t *testing.T) {
	// we start with an empty config
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// this server will respond positively in any case
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.String() == "/v4/info/" {
			w.Write([]byte(`{
			  "general": {
			    "installation_name": "codename",
			    "provider": "aws"
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
			}`))
		} else {
			w.Write([]byte(`{
	      "status_code": 10000,
	      "status_text": "Success",
	      "data": {
	        "Id": "some-test-session-token"
	      }
	    }`))
		}
	}))
	defer mockServer.Close()

	args := loginArguments{}
	args.apiEndpoint = mockServer.URL
	args.email = "email@example.com"
	args.password = "test password"

	result, err := login(args)
	if err != nil {
		t.Error(err)
	}
	if result.email != args.email {
		t.Errorf("Expected '%s', got '%s'", args.email, result.email)
	}
	if result.token != "some-test-session-token" {
		t.Errorf("Expected 'some-test-session-token', got '%s'", result.token)
	}
	if result.alias != "codename" {
		t.Errorf("Expected alias 'codename', got '%s'", result.alias)
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
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// this server will respond positively in any case
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
      "status_code": 10008,
      "status_text": "resource not found"
    }`))
	}))
	defer mockServer.Close()

	args := loginArguments{}
	args.apiEndpoint = mockServer.URL
	args.email = "email@example.com"
	args.password = "bad password"

	_, err = login(args)
	if !IsInvalidCredentialsError(err) {
		t.Errorf("Expected error '%s', got %v", invalidCredentialsError, err)
	}
}

// Test_LoginWhenUserLoggedInBefore simulates an okay login when the user was
// logged in before.
func Test_LoginWhenUserLoggedInBefore(t *testing.T) {
	// this server will respond positively in any case
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
      "status_code": 10000,
      "status_text": "Success",
      "data": {
        "Id": "another-test-session-token"
      }
    }`))
	}))
	defer mockServer.Close()

	// config
	yamlText := `endpoints:
  "` + mockServer.URL + `":
    email: email@foo.com
    token: token
selected_endpoint: "` + mockServer.URL + `"
`
	dir, err := tempConfig(yamlText)
	if err != nil {
		fmt.Printf(yamlText)
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	args := loginArguments{}
	args.apiEndpoint = mockServer.URL
	args.email = "email@example.com"
	args.password = "test password"

	result, loginErr := login(args)
	if loginErr != nil {
		t.Error(err)
	}
	if result.email != args.email {
		t.Errorf("Expected '%s', got '%s'", args.email, result.email)
	}
	if config.Config.Email != args.email {
		t.Errorf("Expected config email to be '%s', got '%s'", args.email, config.Config.Email)
	}
	if result.token != "another-test-session-token" {
		t.Errorf("Expected 'another-test-session-token', got '%s'", result.token)
	}
	if !result.loggedOutBefore {
		t.Error("result.loggedOutBefore was false, expected true")
	}
}
