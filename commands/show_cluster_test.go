package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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
				"release_version": "0.3.0",
				"scaling": {"min": 3, "max": 3},
				"credential_id": "",
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

	details, showErr := getClusterDetails(testArgs.clusterID, showClusterActivityName)
	if showErr != nil {
		t.Error(showErr)
	}

	if details.ID != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, details.ID)
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

	_, err = getClusterDetails(testArgs.clusterID, showClusterActivityName)

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

	_, err = getClusterDetails(testArgs.clusterID, showClusterActivityName)

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

	_, err = getClusterDetails(testArgs.clusterID, showClusterActivityName)

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

// TestShowAWSBYOCCluster tests fetching cluster details for a BYOC cluster on AWS,
// which means the credential_id in cluster details is not empty
func TestShowAWSBYOCCluster(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" && r.URL.Path == "/v4/clusters/cluster-id/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "cluster-id",
				"create_date": "2018-10-25T18:29:34Z",
				"api_endpoint": "https://api.nh9t2.g8s.fra-1.giantswarm.io",
				"owner": "acmeorg",
				"name": "test-cluster",
				"release_version": "4.2.0",
				"scaling": {"min": 2, "max": 2},
				"credential_id": "credential-id",
				"workers": [
					{
						"cpu": {
							"cores": 2
						},
						"memory": {
							"size_gb": 7.5
						},
						"storage": {
							"size_gb": 32
						},
						"aws": {
							"instance_type": "m3.large"
						}
					},
					{
						"cpu": {
							"cores": 2
						},
						"memory": {
							"size_gb": 7.5
						},
						"storage": {
							"size_gb": 32
						},
						"aws": {
							"instance_type": "m3.large"
						}
					}
				]
			}`))
		} else if r.Method == "GET" && r.URL.Path == "/v4/organizations/acmeorg/credentials/credential-id/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "credential-id",
				"provider": "aws",
				"aws": {
					"roles": {
						"admin": "arn:aws:iam::123456789012:role/GiantSwarmAdmin",
						"awsoperator": "arn:aws:iam::123456789012:role/GiantSwarmAWSOperator"
					}
				}
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
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

	details, showErr := getClusterDetails(testArgs.clusterID, showClusterActivityName)
	if showErr != nil {
		t.Error(showErr)
	}

	if details.ID != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, details.ID)
	}

	credentialDetails, err := getOrgCredentials(details.Owner, details.CredentialID, showClusterActivityName)
	if err != nil {
		t.Error(err)
	}

	if credentialDetails != nil && credentialDetails.Aws != nil && credentialDetails.Aws.Roles != nil && credentialDetails.Aws.Roles.Awsoperator == "" {
		t.Error("AWS operator role ARN is empty, should not be.")
	}

	parts := strings.Split(credentialDetails.Aws.Roles.Awsoperator, ":")
	if parts[4] != "123456789012" {
		t.Errorf("Did not get the expected AWS account ID, instead got %s from %s", parts[4], credentialDetails.Aws.Roles.Awsoperator)
	}

}
