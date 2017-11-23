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
		authToken:   "my-token",
	}

	err := verifyShowClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	details, showErr := getClusterDetails(testArgs.clusterID,
		testArgs.authToken, testArgs.apiEndpoint)
	if showErr != nil {
		t.Error(showErr)
	}

	if details.Id != testArgs.clusterID {
		t.Errorf("Expected cluster ID '%s', got '%s'", testArgs.clusterID, details.Id)
	}

}
