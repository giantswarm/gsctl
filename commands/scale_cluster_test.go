package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"

	"github.com/giantswarm/gsctl/config"
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

	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := scaleClusterArguments{
		apiEndpoint: mockServer.URL,
		clusterID:   "cluster-id",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := validateScaleClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if !IsNotLoggedInError(err) {
		t.Error("Expected notLoggedInError, got", err)
	}

}

// TestScaleCluster tests scaling a cluster under normal conditions:
// user logged in.
func TestScaleCluster(t *testing.T) {
	var numWorkersDesired = 5

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			// cluster details before the patch
			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "",
				"api_endpoint": "",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		} else if r.Method == "PATCH" {
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
			workers, _ := patch.S("workers").Children()
			if len(workers) != numWorkersDesired {
				t.Error("Patch request contains", len(workers), "workers, expected 5")
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
		}
	}))
	defer mockServer.Close()

	testArgs := scaleClusterArguments{
		apiEndpoint:       mockServer.URL,
		clusterID:         "cluster-id",
		numWorkersDesired: numWorkersDesired,
	}
	config.Config.Token = "my-token"

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := validateScaleClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	results, scaleErr := scaleCluster(testArgs)
	if scaleErr != nil {
		t.Error(scaleErr)
	}
	if results.numWorkersAfter != testArgs.numWorkersDesired {
		t.Error("Got", results.numWorkersAfter, "workers after scaling, expected", testArgs.numWorkersDesired)
	}

}
