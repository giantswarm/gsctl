package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// TestShowAWSCluster tests fetching cluster details for AWS
func TestShowAWSCluster(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "Name of the cluster",
				"api_endpoint": "https://api.foo.bar",
				"create_date": "2017-11-20T12:00:00.000000Z",
				"owner": "acmeorg",
				"kubernetes_version": "",
        "release_version": "0.3.0",
				"workers": [
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		}
	}))
	defer mockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	details, showErr := getClusterDetails(testArgs.clusterID,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)
	if showErr != nil {
		t.Error(showErr)
	}

	if details.Id != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, details.Id)
	}

}

// TestShowClusterNotAuthorized tests HTTP 401 error handling
func TestShowClusterNotAuthorized(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"code":"INVALID_CREDENTIALS",
				"message":"The requested resource cannot be accessed using the provided credentials. (token not found: unauthenticated)"
			}`))
		}
	}))
	defer mockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-wrong-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, err = getClusterDetails(testArgs.clusterID,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected notAuthorizedError, got nil")
	}

	if !IsNotAuthorizedError(err) {
		t.Errorf("Expected notAuthorizedError, got '%s'", err.Error())
	}
}

// TestShowClusterNotFound tests HTTP 404 error handling
func TestShowClusterNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"code":"RESOURCE_NOT_FOUND",
				"message":"The cluster could not be found."
		}`))
		}
	}))
	defer mockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "non-existing-cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, err = getClusterDetails(testArgs.clusterID,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected clusterNotFoundError, got nil")
	}

	if !IsClusterNotFoundError(err) {
		t.Errorf("Expected clusterNotFoundError, got '%s'", err.Error())
	}
}

// TestShowClusterInternalServerError tests HTTP 500 error handling
func TestShowClusterInternalServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"code":"INTERNAL_ERROR",
				"message":"An unexpected error occurred. Sorry for the inconvenience."
		}`))
		}
	}))
	defer mockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "non-existing-cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, err = getClusterDetails(testArgs.clusterID,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected internalServerError, got nil")
	}

	if !IsInternalServerError(err) {
		t.Errorf("Expected internalServerError, got '%s'", err.Error())
	}
}

// TestShowClusterNotLoggedIn tests the case where the client is not logged in
func TestShowClusterNotLoggedIn(t *testing.T) {
	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: "foo.bar",
		clusterID:   "cluster-id",
		authToken:   "",
	}

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if !IsNotLoggedInError(err) {
		t.Errorf("Expected notLoggedInError, got '%s'", err.Error())
	}

}

// TestShowClusterMissingID tests the case where the cluster ID is missing
func TestShowClusterMissingID(t *testing.T) {
	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showClusterArguments{
		apiEndpoint: "foo.bar",
		clusterID:   "",
		authToken:   "auth-token",
	}

	err := verifyShowClusterPreconditions(testArgs, []string{})
	if !IsClusterIDMissingError(err) {
		t.Errorf("Expected clusterIdMissingError, got '%s'", err.Error())
	}

}
