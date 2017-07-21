package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/spf13/viper"

	"github.com/giantswarm/gsctl/config"
)

// TestScaleClusterNotLoggedIn tests if e can prevent an attempt to do things
// when not logged in and no token has been provided.
func TestScaleClusterNotLoggedIn(t *testing.T) {
	defer viper.Reset()

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

	err := verifyScaleClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if !IsNotLoggedInError(err) {
		t.Error("Expected notLoggedInError, got", err)
	}

}

// TestScaleCluster tests scaling a cluster under normal conditions:
// user logged in,
// The test is more invoved than it should be, as the API currently
// does not return cluster details with the PATCH response.
// See https://github.com/giantswarm/api/issues/437
func TestScaleCluster(t *testing.T) {
	defer viper.Reset()

	var numWorkersDesired = 5
	var requestCount = 0

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clusterDetailsJSON := []byte(`{
			"id": "cluster-id",
			"name": "",
			"api_endpoint": "",
			"create_date": "2017-05-16T09:30:31.192170835Z",
			"owner": "acmeorg",
			"kubernetes_version": "",
			"workers": [
				{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
				{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
				{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
			]
		}`)

		// modify response for the second GET request
		if requestCount > 1 {
			clusterDetailsJSON = []byte(`{
				"id": "cluster-id",
				"name": "",
				"api_endpoint": "",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"kubernetes_version": "",
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			// cluster details before the patch
			requestCount++
			t.Log("requestCount (GET):", requestCount)
			w.Write(clusterDetailsJSON)
		} else if r.Method == "PATCH" {
			requestCount++
			t.Log("requestCount (PATCH):", requestCount)
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

			w.Write(clusterDetailsJSON)
		}
	}))
	defer mockServer.Close()

	testArgs := scaleClusterArguments{
		apiEndpoint:       mockServer.URL,
		clusterID:         "cluster-id",
		numWorkersDesired: numWorkersDesired,
	}
	config.Config.Token = "my-token"

	err := verifyScaleClusterPreconditions(testArgs, []string{testArgs.clusterID})
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
