package client

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
)

// TestRedactPasswordArgs tests redactPasswordArgs().
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
// server connection.
func TestV2NoConnection(t *testing.T) { // Our test server.

	// a non-existing endpoint (must use an IP, not a hostname)
	config := &Configuration{
		Endpoint: "http://127.0.0.1:55555",
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password", nil)

	if err == nil {
		t.Error("Expected 'connection refused' error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	_, ok = clientAPIError.OriginalError.(*net.OpError)
	if !ok {
		t.Error("Type assertion to *net.OpError not successful")
	}

	t.Logf("clientAPIError: %#v", clientAPIError)

	if clientAPIError.ErrorMessage == "" {
		t.Error("ErrorMessage was empty, expected helpful message.")
	}
	if clientAPIError.ErrorDetails == "" {
		t.Error("ErrorDetails was empty, expected helpful message.")
	}
}

// TestV2HostnameUnresolvable checks out how the latest client deals with a
// non-resolvable host name.
func TestV2HostnameUnresolvable(t *testing.T) { // Our test server.

	// a non-existing host name
	config := &Configuration{
		Endpoint: "http://non.existing.host.name",
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password", nil)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	_, ok = clientAPIError.OriginalError.(*net.DNSError)
	if !ok {
		t.Error("Type assertion to *net.DNSError not successful")
	}

	t.Logf("clientAPIError: %#v", clientAPIError)

	if clientAPIError.ErrorMessage == "" {
		t.Error("ErrorMessage was empty, expected helpful message.")
	}
	if clientAPIError.ErrorDetails == "" {
		t.Error("ErrorDetails was empty, expected helpful message.")
	}
}

// TestV2Timeout tests if the latest client handles timeouts as expected.
func TestV2Timeout(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce a timeout longer than the client's
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	clientConfig := &Configuration{
		Endpoint: ts.URL,
		Timeout:  500 * time.Millisecond,
	}
	gsClient, err := NewV2(clientConfig)
	if err != nil {
		t.Error(err)
	}
	resp, err := gsClient.CreateAuthToken("email", "password", nil)
	if err == nil {
		t.Error("Expected Timeout error, got nil")
		t.Logf("resp: %#v", resp)
	} else {
		clientAPIError, ok := err.(*clienterror.APIError)
		if !ok {
			t.Error("Type assertion err.(*clienterror.APIError) not successful")
		}
		if !clientAPIError.IsTimeout {
			t.Error("Expected clientAPIError.IsTimeout to be true, got false")
		}
	}
}

// TestV2UserAgent tests whether our user-agent header appears in requests.
func TestV2UserAgent(t *testing.T) {
	clientConfig := &Configuration{
		UserAgent: "my own user agent/1.0",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("user-agent")

		if ua != clientConfig.UserAgent {
			t.Errorf("Expected '%s', got '%s'", clientConfig.UserAgent, ua)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "NONE", "message": "none"}`))
	}))
	defer ts.Close()

	clientConfig.Endpoint = ts.URL

	gsClient, err := NewV2(clientConfig)
	if err != nil {
		t.Error(err)
	}

	// just issue a request, don't care about the result.
	gsClient.CreateAuthToken("email", "password", nil)
}

// TestV2Forbidden tests out how the latest client gives access to
// HTTP error details for a 403 error.
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

	response, err := gsClient.CreateAuthToken("email", "password", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if response != nil {
		t.Error("Expected nil response")
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	if clientAPIError.HTTPStatusCode != http.StatusForbidden {
		t.Error("Expected HTTP status 403, got", clientAPIError.HTTPStatusCode)
	}
}

// TestV2Unauthorized tests out how the latest client gives access to
// HTTP error details for a 401 error.
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

	_, err = gsClient.DeleteAuthToken("foo", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	t.Logf("err: %#v", err)

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	if clientAPIError.HTTPStatusCode != http.StatusUnauthorized {
		t.Error("Expected HTTP status 401, got", clientAPIError.HTTPStatusCode)
	}
}

// TestV2AuxiliaryParams checks whether the client carries through our auxiliary
// parameters.
func TestV2AuxiliaryParams(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-Request-ID") != "request-id" {
			t.Error("Header X-Request-ID not available")
		}
		if r.Header.Get("X-Giant-Swarm-CmdLine") != "command-line" {
			t.Error("Header X-Giant-Swarm-CmdLine not available")
		}
		if r.Header.Get("X-Giant-Swarm-Activity") != "activity-name" {
			t.Error("Header X-Giant-Swarm-Activity not available")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"foo": "bar"}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	ap := gsClient.DefaultAuxiliaryParams()
	ap.RequestID = "request-id"
	ap.CommandLine = "command-line"
	ap.ActivityName = "activity-name"

	gsClient.CreateAuthToken("foo", "bar", ap)
}

// TestV2CreateAuthToken checks out how creating an auth token works in
// our new client.
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

	response, err := gsClient.CreateAuthToken("foo", "bar", nil)
	if err != nil {
		t.Error(err)
	}

	if response.Payload.AuthToken != "e5239484-2299-41df-b901-d0568db7e3f9" {
		t.Errorf("Didn't get the expected token. Got %s", response.Payload.AuthToken)
	}
}

// TestV2DeleteAuthToken checks out how to issue an authenticted request
// using the new client.
func TestV2DeleteAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "giantswarm test-token" {
			t.Error("Bad authorization header:", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been successfully deleted."}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	response, err := gsClient.DeleteAuthToken("test-token", nil)
	if err != nil {
		t.Error(err)
	}

	if response.Payload.Code != "RESOURCE_DELETED" {
		t.Errorf("Didn't get the RESOURCE_DELETED message. Got '%s'", response.Payload.Code)
	}
}

func TestGetClusterStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cluster": {
				"conditions": [
					{
						"status": "True",
						"type": "Created"
					}
				],
				"network": {
					"cidr": ""
				},
				"nodes": [
					{
						"name": "4jr2w-master-000000",
						"version": "2.0.1"
					},
					{
						"name": "4jr2w-worker-000001",
						"version": "2.0.1"
					}
				],
				"resources": [],
				"versions": [
					{
						"date": "0001-01-01T00:00:00Z",
						"semver": "2.0.1"
					}
				]
			}
		}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	status, err := gsClient.GetClusterStatus("cluster-id", nil)
	if err != nil {
		t.Error(err)
	}

	if len(status.Cluster.Nodes) != 2 {
		t.Errorf("Expected status.Nodes to have length 2, but has %d. status: %#v", len(status.Cluster.Nodes), status)
	}
}

func TestGetClusterStatusEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cluster": {
				"nodes": []
			}
		}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := NewV2(config)
	if err != nil {
		t.Error(err)
	}

	status, err := gsClient.GetClusterStatus("cluster-id", nil)
	if err != nil {
		t.Error(err)
	}

	if len(status.Cluster.Nodes) != 0 {
		t.Errorf("Expected status.Nodes to have length 0. Has length %d", len(status.Cluster.Nodes))
	}
}

// Test_GetDefaultCluster tests the GetDefaultCluster function
// for the case that only one cluster exists
func Test_GetDefaultCluster(t *testing.T) {
	// returns one cluster
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
      {
        "create_date": "2017-04-16T09:30:31.192170835Z",
        "id": "cluster-id",
        "name": "Some random test cluster",
				"owner": "acme"
      }
    ]`))
	}))
	defer mockServer.Close()

	// config
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL
	dir, err := config.TempConfig(yamlText)
	defer os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}

	clientV2, err := NewWithConfig(mockServer.URL, "")

	clusterID, err := clientV2.GetDefaultCluster(nil)
	if err != nil {
		t.Error(err)
	}
	if clusterID != "cluster-id" {
		t.Errorf("Expected 'cluster-id', got %#v", clusterID)
	}
}
