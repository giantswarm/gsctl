package cluster

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	models "github.com/giantswarm/gsclientgen/models"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// TestShowAWSClusterV4 tests fetching V4 cluster details for AWS,
// for a cluster that does not have BYOC credentials and no status yet.
func TestShowAWSClusterV4(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v4/clusters/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": "cluster-id",
						"name": "Name of the cluster",
						"owner": "acme"
					}
				]`))

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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
		verbose:         true,
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	if detailsV4.ID != testArgs.clusterNameOrID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterNameOrID, detailsV4.ID)
	}

}

// TestShowAWSClusterV5 tests fetching V4 cluster details for AWS,
// for a cluster that does not have BYOC credentials.
func TestShowAWSClusterV5(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v4/clusters/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": "cluster-id",
						"name": "Name of the cluster",
						"owner": "acme"
					}
				]`))

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
					"master_nodes": {
						"availability_zones": ["eu-west-1b", "eu-west-1c", "eu-west-1d"],
						"high_availability": true,
						"num_ready": 3
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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
		verbose:         true,
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	if detailsV5.ID != testArgs.clusterNameOrID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterNameOrID, detailsV5.ID)
	}

}

// TestShowAWSClusterV5NoHAMasters tests fetching V4 cluster details for AWS,
// without HA Masters support.
func TestShowAWSClusterV5NoHAMasters(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v4/clusters/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": "cluster-id",
						"name": "Name of the cluster",
						"owner": "acme"
					}
				]`))

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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
		verbose:         true,
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	if detailsV5.ID != testArgs.clusterNameOrID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterNameOrID, detailsV5.ID)
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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-wrong-token",
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
	if err != nil {
		t.Error(err)
	}

	_, err = getClusterDetailsV4(testArgs)

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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "non-existing-cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "non-existing-cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	testArgs := Arguments{
		apiEndpoint:     "foo.bar",
		clusterNameOrID: "cluster-id",
		authToken:       "",
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
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

	testArgs := Arguments{
		apiEndpoint:     "foo.bar",
		clusterNameOrID: "",
		authToken:       "auth-token",
	}

	err := verifyPreconditions(testArgs, []string{})
	if !errors.IsClusterNameOrIDMissingError(err) {
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
			case "/v4/clusters/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": "cluster-id",
						"name": "Name of the cluster",
						"owner": "acme"
					}
				]`))

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

	testArgs := Arguments{
		apiEndpoint:     mockServer.URL,
		clusterNameOrID: "cluster-id",
		scheme:          "giantswarm",
		authToken:       "my-token",
	}

	err := verifyPreconditions(testArgs, []string{testArgs.clusterNameOrID})
	if err != nil {
		t.Error(err)
	}

	detailsV4, _, _, _, credentialDetails, showErr := getClusterDetails(testArgs)
	if showErr != nil {
		t.Error(showErr)
	}

	if detailsV4.ID != testArgs.clusterNameOrID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterNameOrID, detailsV4.ID)
	}

	if credentialDetails != nil && credentialDetails.Aws != nil && credentialDetails.Aws.Roles != nil && credentialDetails.Aws.Roles.Awsoperator == "" {
		t.Error("AWS operator role ARN is empty, should not be.")
	}

	parts := strings.Split(credentialDetails.Aws.Roles.Awsoperator, ":")
	if parts[4] != "123456789012" {
		t.Errorf("Did not get the expected AWS account ID, instead got %s from %s", parts[4], credentialDetails.Aws.Roles.Awsoperator)
	}

}

func TestFormatClusterLabels(t *testing.T) {
	mockLabels := map[string]string{
		"shouldbeignored.giantswarm.io": "imhidden",
		"release.giantswarm.io":         "0.0.0",
	}

	result := formatClusterLabels(mockLabels)

	if len(result) != 1 {
		t.Errorf("formatted cluster labels result has invalid length. Expected %d got %d", 1, len(result))
	}

	if result[0] != "Labels:|-" {
		t.Errorf("formatted cluster labels result expected '%s' got '%s'", "Labels:|-", result[0])
	}

	mockLabels["testkey"] = "testvalue"
	mockLabels["veryvalidkey"] = "veryvalidvalue"

	result = formatClusterLabels(mockLabels)

	if len(result) != 2 {
		t.Errorf("formatted cluster labels result has invalid length. Expected %d got %d", 2, len(result))
	}

	if !strings.HasPrefix(result[0], "Labels:|") {
		t.Errorf("formatted cluster labels result expected '%s' to start with '%s'", result[0], "Labels:|")
	}

	if !strings.HasPrefix(result[1], "|") {
		t.Errorf("formatted cluster labels result expected '%s' to start with '%s'", result[1], "|")
	}
}

func Test_FormatMasterNodes(t *testing.T) {
	testCases := []struct {
		model                     *models.V5ClusterDetailsResponseMasterNodes
		expectedAvailabilityZones string
		expectedNumOfReadyNodes   string
	}{
		{
			model:                     nil,
			expectedAvailabilityZones: "n/a",
			expectedNumOfReadyNodes:   "n/a",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{},
				HighAvailability:  false,
				NumReady:          toInt8Ptr(0),
			},
			expectedAvailabilityZones: "n/a",
			expectedNumOfReadyNodes:   "0",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{"some-zone"},
				HighAvailability:  false,
				NumReady:          toInt8Ptr(1),
			},
			expectedAvailabilityZones: "some-zone",
			expectedNumOfReadyNodes:   "1",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{"some-zone-a", "some-zone-b", "some-zone-c"},
				HighAvailability:  false,
				NumReady:          toInt8Ptr(3),
			},
			expectedAvailabilityZones: "some-zone-a, some-zone-b, some-zone-c",
			expectedNumOfReadyNodes:   "3",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{"some-zone-a", "some-zone-b", "some-zone-c"},
				HighAvailability:  false,
				NumReady:          toInt8Ptr(0),
			},
			expectedAvailabilityZones: "some-zone-a, some-zone-b, some-zone-c",
			expectedNumOfReadyNodes:   "0",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{"some-zone-a", "some-zone-b", "some-zone-c"},
				HighAvailability:  false,
				NumReady:          nil,
			},
			expectedAvailabilityZones: "some-zone-a, some-zone-b, some-zone-c",
			expectedNumOfReadyNodes:   "n/a",
		},
		{
			model: &models.V5ClusterDetailsResponseMasterNodes{
				AvailabilityZones: []string{"some-zone-a", "some-zone-b", "some-zone-c"},
				HighAvailability:  false,
				NumReady:          nil,
			},
			expectedAvailabilityZones: "some-zone-a, some-zone-b, some-zone-c",
			expectedNumOfReadyNodes:   "n/a",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			azs, numOfReadyNodes := formatMasterNodes(tc.model)

			if azs != tc.expectedAvailabilityZones {
				t.Errorf("Case %d - Result did not match.\nExpected: %s, Got: %s", i, tc.expectedAvailabilityZones, azs)
			} else if numOfReadyNodes != tc.expectedNumOfReadyNodes {
				t.Errorf("Case %d - Result did not match.\nExpected: %s, Got: %s", i, tc.expectedNumOfReadyNodes, numOfReadyNodes)
			}
		})
	}
}

func toInt8Ptr(i int) *int8 {
	c := int8(i)

	return &c
}
