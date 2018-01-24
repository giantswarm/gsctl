package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/giantswarm/gsctl/config"
)

var successorVersionTests = []struct {
	myVersion        string
	allVersions      []string
	successorVersion string
}{
	{
		"1.2.3",
		[]string{"1.2.3", "1.2.4", "1.2.5", "10.2.3", "2.0.0", "0.1.0"},
		"1.2.4",
	},
	// none of the versions in the slice is higher
	{
		"4.5.2",
		[]string{"3.2.1", "0.5.1", "0.5.0", "0.6.0", "4.5.2"},
		"",
	},
}

func TestSuccessorReleaseVersion(t *testing.T) {
	for i, tc := range successorVersionTests {
		v := successorReleaseVersion(tc.myVersion, tc.allVersions)
		if v != tc.successorVersion {
			t.Errorf("%d. Expected %s, got %s", i, tc.successorVersion, v)
		}
	}
}

func TestUpgradeCluster(t *testing.T) {
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" && r.RequestURI == "/v4/clusters/cluster-id/" {
			// cluster details before the patch
			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "This is my cluster name",
				"api_endpoint": "http://test.api.endpoint",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"release_version": "1.2.3",
				"kubernetes_version": "",
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		} else if r.RequestURI == "/v4/releases/" {
			// return list of releases
			w.Write([]byte(`[
			  {
					"timestamp": "2017-10-15T12:00:00Z",
			    "version": "0.1.0",
			    "active": true,
			    "changelog": [
			      {"component": "vault", "description": "Foo"}
			    ],
			    "components": [
			      {"name": "kubernetes", "version": "1.8.0"}
			    ]
			  },
				{
					"timestamp": "2017-10-15T12:00:00Z",
			    "version": "1.2.5",
			    "active": true,
			    "changelog": [
			      {"component": "vault", "description": "Bar"}
			    ],
			    "components": [
			      {"name": "kubernetes", "version": "1.8.0"}
			    ]
			  }
			]`))
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
			if !patch.Exists("release_version") {
				t.Error("Patch request body does not contain 'release_version' key.")
			}

			version := patch.Path("release_version").Data()
			if version != "1.2.5" {
				t.Errorf("Patch request contained version %s, expected 1.2.5", version)
			}

			w.Write([]byte(`{
				"id": "cluster-id",
				"name": "This is my cluster name",
				"api_endpoint": "http://test.api.endpoint",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg",
				"release_version": "1.2.5",
				"kubernetes_version": "",
				"workers": [
					{"memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		}
	}))
	defer mockServer.Close()

	testArgs := upgradeClusterArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "my-token",
		clusterID:   "cluster-id",
		force:       true,
	}

	err := verifyUpgradeClusterPreconditions(testArgs, []string{testArgs.clusterID})
	if err != nil {
		t.Error(err)
	}

	results, upgradeErr := upgradeCluster(testArgs)
	if upgradeErr != nil {
		t.Error(upgradeErr)
	}
	if results.versionAfter != "1.2.5" {
		t.Error("Got version", results.versionAfter, ", expected 1.2.5")
	}

}
