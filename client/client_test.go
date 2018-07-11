package client

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/go-openapi/runtime"
)

func TestTimeout(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce a timeout longer than the client's
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	clientConfig := Configuration{
		Endpoint: ts.URL,
		Timeout:  1 * time.Second,
	}
	apiClient, clientErr := New(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	_, _, err := apiClient.GetUserOrganizations("foo")
	if err == nil {
		t.Error("Expected Timeout error, got nil")
	} else {
		if err, ok := err.(net.Error); ok && !err.Timeout() {
			t.Error("Expected Timeout error, got", err)
		}
	}
}

// TestUserAgent tests whether request have the proper User-Agent header
// and if ParseGenericResponse works as expected
func TestUserAgent(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return valid JSON containing user agent string received
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": "BAD_REQUEST", "message": "user-agent: ` + r.Header.Get("User-Agent") + `"}`))
	}))
	defer ts.Close()

	clientConfig := Configuration{
		Endpoint:  ts.URL,
		UserAgent: "my own user agent/1.0",
	}
	apiClient, clientErr := New(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	// We use GetUserOrganizations just to issue a request. We could use any other
	// API call, it wouldn't matter.
	_, apiResponse, _ := apiClient.GetUserOrganizations("foo")

	gr, err := ParseGenericResponse(apiResponse.Payload)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(gr.Message, clientConfig.UserAgent) {
		t.Error("UserAgent string could not be found")
	}
}

// TestRedactPasswordArgs tests redactPasswordArgs()
func TestRedactPasswordArgs(t *testing.T) {
	argtests := []struct {
		in  string
		out string
	}{
		// these remain unchangd
		{"foo", "foo"},
		{"foo bar", "foo bar"},
		{"foo bar blah", "foo bar blah"},
		{"foo bar blah -p mypass", "foo bar blah -p mypass"},
		{"foo bar blah -p=mypass", "foo bar blah -p=mypass"},
		// these will be altered
		{"foo bar blah --password mypass", "foo bar blah --password REDACTED"},
		{"foo bar blah --password=mypass", "foo bar blah --password=REDACTED"},
		{"foo login blah -p mypass", "foo login blah -p REDACTED"},
		{"foo login blah -p=mypass", "foo login blah -p=REDACTED"},
	}

	for _, tt := range argtests {
		in := strings.Split(tt.in, " ")
		out := strings.Join(redactPasswordArgs(in), " ")
		if out != tt.out {
			t.Errorf("want '%q', have '%s'", tt.in, tt.out)
		}
	}
}

// TestV2NoConnection checks out how the latest client deals with a missing
// server connection
func TestV2NoConnection(t *testing.T) { // Our test server.

	// a non-existing endpoint (must use an IP, not a hostname)
	config := &Configuration{
		Endpoint: "http://127.0.0.1:55555",
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password")

	if err == nil {
		t.Error("Expected 'connection refused' error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}
	rootError, ok := microerror.Cause(err).(*url.Error)
	if !ok {
		t.Error("Type assertion to *url.Error failed")
	}

	t.Logf("%#v", rootError.Err)
}

// TestV2NoHostnameUnresolvable checks out how the latest client deals with a
// non-resolvable host name
func TestV2NoHostnameUnresolvable(t *testing.T) { // Our test server.

	// a non-existing host name
	config := &Configuration{
		Endpoint: "http://non.existing.host.name",
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}
	urlError, ok := microerror.Cause(err).(*url.Error)
	if !ok {
		t.Error("Type assertion to *url.Error failed")
	}

	netOpError, ok := urlError.Err.(*net.OpError)
	if !ok {
		t.Error("Type assertion to *net.OpError failed")
	}

	rootError, ok := netOpError.Err.(*net.DNSError)
	if !ok {
		t.Error("Type assertion to *net.DNSError failed")
	}

	if rootError.Err != "no such host" {
		t.Errorf("Expected DNS error 'no such host', got '%s'", rootError.Err)
	}

}

// TestV2Forbidden tests out how the latest client gives access to
// HTTP error details for a 403 error
func TestV2Forbidden(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`Access forbidden`))
	}))
	defer ts.Close()

	gsClient, err := NewV2(&Configuration{Endpoint: ts.URL})
	if err != nil {
		t.Error(err)
	}

	_, err = gsClient.CreateAuthToken("email", "password")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	e, ok := microerror.Cause(err).(*runtime.APIError)
	if !ok {
		t.Error("Type assertion to *runtime.APIError failed")
	}

	if e.Code != 403 {
		t.Error("Expected status code 403, got", e.Code)
	}
}

// TestV2Unauthorized tests out how the latest client gives access to
// HTTP error details for a 401 error
func TestV2Unauthorized(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "PERMISSION_DENIED", "message": "Not authorized"}`))
	}))
	defer ts.Close()

	gsClient, err := NewV2(&Configuration{Endpoint: ts.URL})
	if err != nil {
		t.Error(err)
	}

	_, err = gsClient.DeleteAuthToken("foo")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	t.Logf("err: %#v", err)

	// e, ok := microerror.Cause(err).(*runtime.APIError)
	// if !ok {
	// 	t.Error("Type assertion to *runtime.APIError failed")
	// }

	// if e.Code != 403 {
	// 	t.Error("Expected status code 403, got", e.Code)
	// }
}

// TestV2CreateAuthToken checks out how creating an auth token works in
// our new client
func TestV2CreateAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth_token": "e5239484-2299-41df-b901-d0568db7e3f9"}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	if responseBody.AuthToken != "e5239484-2299-41df-b901-d0568db7e3f9" {
		t.Errorf("Didn't get the expected token. Got %s", responseBody.AuthToken)
	}
}

// TestV2DeleteAuthToken checks out how to issue an authenticted request
// using the new client
func TestV2DeleteAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "giantswarm test-token" {
			t.Error("Bad authorization header:", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been succesfully deleted."}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.DeleteAuthToken("test-token")
	if err != nil {
		t.Error(err)
	}

	if responseBody.Code != "RESOURCE_DELETED" {
		t.Errorf("Didn't get the RESOURCE_DELETED message. Got '%s'", responseBody.Code)
	}
}
