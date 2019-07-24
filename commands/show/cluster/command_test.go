package cluster

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

// TestShowAWSClusterV4 tests fetching V4 cluster details for AWS,
// for a cluster that does not have BYOC credentials and no status yet.
func TestShowAWSClusterV4(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v4/clusters/cluster-id/":
				w.WriteHeader(http.StatusOK)
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
			case "/v4/clusters/cluster-id/status/":
				// simulating the case where cluster status is not yet available,
				// to keep it simple here
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
			case "/v5/clusters/cluster-id/":
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Cluster does not exist or is not accessible."}`))
			}
		}
	}))
	defer mockServer.Close()

	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
		verbose:     true,
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	detailsV4, detailsV5, _, statusV4, credentials, err := getClusterDetails(testArgs)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if detailsV5 != nil {
		t.Errorf("Expected detailsV5 to be nil, got %#v", detailsV5)
	}

	if statusV4 != nil {
		t.Errorf("Expected statusV4 to be nil, got %v", statusV4)
	}

	if credentials != nil {
		t.Errorf("Expected credentials to be nil, got %v", credentials)
	}

	if detailsV4 == nil {
		t.Fatal("Expected V4 cluster details, got nil")
	}

	if detailsV4.ID != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, detailsV4.ID)
	}

}

// TestShowAWSClusterV5 tests fetching V4 cluster details for AWS,
// for a cluster that does not have BYOC credentials.
func TestShowAWSClusterV5(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v5/clusters/cluster-id/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "cluster-id",
					"name": "AWS v5 cluster",
					"api_endpoint": "https://api.foo.bar",
					"create_date": "2019-07-09T12:00:00.000000Z",
					"owner": "acmeorg",
					"release_version": "9.1.2",
					"credential_id": "",
					"master": {
						"availability_zone": "eu-west-1d"
					}
				}`))
			case "/v5/clusters/cluster-id/nodepools/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": "a7r",
						"name": "Node pool name",
						"availability_zones": [
							"eu-west-1d"
						],
						"scaling": {
							"min": 2,
							"max": 5
						},
						"node_spec": {
							"aws": {
								"instance_type": "p3.8xlarge"
							},
							"volume_sizes_gb": {
								"docker": 100,
								"kubelet": 100
							}
						},
						"status": {
							"nodes": 2,
							"nodes_ready": 2
						}
					}
				]`))
			case "/v4/clusters/cluster-id/":
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Cluster does not exist or is not accessible."}`))
			default:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "We do this to notice any unexpected endpoint being called."}`))
			}
		}
	}))
	defer mockServer.Close()

	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
		verbose:     true,
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	detailsV4, detailsV5, _, statusV4, credentials, err := getClusterDetails(testArgs)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if detailsV4 != nil {
		t.Errorf("Expected detailsV4 to be nil, got %#v", detailsV5)
	}

	if statusV4 != nil {
		t.Errorf("Expected statusV4 to be nil, got %v", statusV4)
	}

	if credentials != nil {
		t.Errorf("Expected credentials to be nil, got %v", credentials)
	}

	if detailsV5 == nil {
		t.Fatal("Expected V5 cluster details, got nil")
	}

	if detailsV5.ID != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, detailsV5.ID)
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
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-wrong-token",
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, err = getClusterDetailsV4(testArgs.clusterID)

	if err == nil {
		t.Fatal("Expected NotAuthorizedError, got nil")
	}

	if !errors.IsNotAuthorizedError(err) {
		t.Errorf("Expected NotAuthorizedError, got '%s'", err.Error())
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
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "non-existing-cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, _, _, _, _, err = getClusterDetails(testArgs)
	if err == nil {
		t.Fatal("Expected ClusterNotFoundError, got nil")
	}

	if !errors.IsClusterNotFoundError(err) {
		t.Errorf("Expected ClusterNotFoundError, got '%s'", err.Error())
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
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "non-existing-cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	_, _, _, _, _, err = getClusterDetails(testArgs)
	if err == nil {
		t.Fatal("Expected InternalServerError, got nil")
	}

	if !errors.IsInternalServerError(err) {
		t.Errorf("Expected InternalServerError, got '%s'", err.Error())
	}
}

// TestShowClusterNotLoggedIn tests the case where the client is not logged in
func TestShowClusterNotLoggedIn(t *testing.T) {
	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: "foo.bar",
		clusterID:   "cluster-id",
		authToken:   "",
	}

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if !errors.IsNotLoggedInError(err) {
		t.Errorf("Expected NotLoggedInError, got '%s'", err.Error())
	}

}

// TestShowClusterMissingID tests the case where the cluster ID is missing
func TestShowClusterMissingID(t *testing.T) {
	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: "foo.bar",
		clusterID:   "",
		authToken:   "auth-token",
	}

	err := verifyShowClusterPreconditions(testArgs, []string{})
	if !errors.IsClusterIDMissingError(err) {
		t.Errorf("Expected clusterIdMissingError, got '%s'", err.Error())
	}

}

// TestShowAWSBYOCCluster tests fetching cluster details for a BYOC cluster on AWS,
// which means the credential_id in cluster details is not empty
func TestShowAWSBYOCClusterV4(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			switch r.URL.Path {
			case "/v4/clusters/cluster-id/":
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
			case "/v4/organizations/acmeorg/credentials/credential-id/":
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
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
	defer mockServer.Close()

	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := showClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		scheme:      "giantswarm",
		authToken:   "my-token",
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	detailsV4, _, _, _, credentialDetails, showErr := getClusterDetails(testArgs)
	if showErr != nil {
		t.Error(showErr)
	}

	if detailsV4.ID != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, detailsV4.ID)
	}

	if credentialDetails != nil && credentialDetails.Aws != nil && credentialDetails.Aws.Roles != nil && credentialDetails.Aws.Roles.Awsoperator == "" {
		t.Error("AWS operator role ARN is empty, should not be.")
	}

	parts := strings.Split(credentialDetails.Aws.Roles.Awsoperator, ":")
	if parts[4] != "123456789012" {
		t.Errorf("Did not get the expected AWS account ID, instead got %s from %s", parts[4], credentialDetails.Aws.Roles.Awsoperator)
	}

}
