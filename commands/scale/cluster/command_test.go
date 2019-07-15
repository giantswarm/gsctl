package cluster

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

// TestScaleClusterNotLoggedIn tests if we can prevent an attempt to do things
// when not logged in and no token has been provided.
func TestScaleClusterNotLoggedIn(t *testing.T) {
	// This server should not get any request, because we avoid unauthenticated requests.
	// That's why it issues an error in case it does.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("TestScaleClusterNotLoggedIn mockServer request:", r.Method, r.URL)
		t.Error("A request has been sent although we don't have a token.")
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := scaleClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
	}

	flags.CmdAPIEndpoint = mockServer.URL

	err := validateScaleCluster(testArgs, []string{testArgs.clusterID}, 5, 5, 5)
	if !errors.IsNotLoggedInError(err) {
		t.Error("Expected NotLoggedInError, got", err)
	}

}

// TestScaleCluster tests scaling a cluster under normal conditions:
// user logged in.
func TestScaleCluster(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" && r.URL.String() == "/v4/clusters/cluster-id/" {
			// cluster details before the patch
			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "",
				"api_endpoint": "",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"scaling": {
					"max":3,
					"min":3
				},
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		} else if r.Method == "PATCH" && r.URL.String() == "/v4/clusters/cluster-id/" {
			// inspect PATCH request body
			patchBytes, readErr := ioutil.ReadAll(r.Body)
			if readErr != nil {
				t.Error(readErr)
			}
			patch, parseErr := gabs.ParseJSON(patchBytes)
			if parseErr != nil {
				t.Error(parseErr)
			}
			if !patch.Exists("workers") {
				t.Error("Patch request body does not contain 'workers' key.")
			}

			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "",
				"api_endpoint": "",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		} else if r.Method == "GET" && r.URL.String() == "/v4/clusters/cluster-id/status/" {

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
					"scaling":{
						"desiredCapacity": 3
					},					
					"versions": [
						{
							"date": "0001-01-01T00:00:00Z",
							"semver": "2.0.1"
						}
					]
				}
			}`))
		}
	}))
	defer mockServer.Close()

	// config
	yamlText := `
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
    provider: aws
selected_endpoint: ` + mockServer.URL
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
	if err != nil {
		t.Error(err)
	}

	testArgs := scaleClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
		workersMax:  int64(5),
		workersMin:  int64(5),
	}
	config.Config.Token = "my-token"

	flags.CmdAPIEndpoint = mockServer.URL

	err = validateScaleCluster(testArgs, []string{testArgs.clusterID}, 3, 3, 3)
	if err != nil {
		t.Error(err)
	}

	status, err := getClusterStatus(testArgs.clusterID, "scale-cluster")
	if err != nil {
		t.Error(err)
	}

	if status.Cluster.Scaling.DesiredCapacity != 3 {
		t.Errorf("Expected status.Scaling.DesiredCapacity to be 3, but is %d. status: %#v", status.Cluster.Scaling.DesiredCapacity, status)
	}

	_, scaleErr := scaleCluster(testArgs)

	if scaleErr != nil {
		t.Error(scaleErr)
	}
}
