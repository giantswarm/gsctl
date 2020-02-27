package cluster

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/giantswarm/gscliauth/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

func Test_successorReleaseVersion(t *testing.T) {
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

	for i, tc := range successorVersionTests {
		v := successorReleaseVersion(tc.myVersion, tc.allVersions)
		if v != tc.successorVersion {
			t.Errorf("%d. Expected %s, got %s", i, tc.successorVersion, v)
		}
	}
}

func TestUpgradeCluster(t *testing.T) {
	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.RequestURI == "/v4/clusters/cluster-id/" {
			// cluster details before the patch
			w.WriteHeader(http.StatusOK)
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
			w.WriteHeader(http.StatusOK)
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

			w.WriteHeader(http.StatusOK)
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
		} else {
			t.Logf("Mock sevrer requst to %s %s", r.Method, r.RequestURI)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Not found"}`))
		}
	}))
	defer mockServer.Close()

	testArgs := Arguments{
		APIEndpoint: mockServer.URL,
		AuthToken:   "my-token",
		ClusterID:   "cluster-id",
		Force:       true,
	}

	err := validateUpgradeClusterPreconditions(testArgs, []string{testArgs.ClusterID})
	if err != nil {
		t.Error(err)
	}

	results, upgradeErr := upgradeCluster(testArgs)
	if upgradeErr != nil {
		t.Fatal(upgradeErr)
	}

	if results.versionAfter != "1.2.5" {
		t.Error("Got version", results.versionAfter, ", expected 1.2.5")
	}
}

func Test_collectArguments(t *testing.T) {
	tests := []struct {
		name                string
		positionalArguments []string
		commandExecution    func()
		resultingArgs       Arguments
	}{
		{
			name:                "Test 1: minimal arguments",
			positionalArguments: []string{"clusterid"},
			commandExecution: func() {
				initFlags()
				Command.ParseFlags([]string{
					"clusterid",
					"--force",
				})
			},
			resultingArgs: Arguments{
				ClusterID: "clusterid",
				Force:     true,
			},
		},
	}

	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// temp config
			fs := afero.NewMemMapFs()
			configDir := testutils.TempDir(fs)
			config.Initialize(fs, configDir)

			initFlags()
			tt.commandExecution()

			got := collectArguments(tt.positionalArguments)

			if diff := cmp.Diff(tt.resultingArgs, got, nil); diff != "" {
				t.Errorf("Test %d - Resulting args unequal. (-expected +got):\n%s", (index + 1), diff)
			}
		})
	}
}

func Test_validateUpgradeClusterPreconditions(t *testing.T) {
	type args struct {
		args        Arguments
		cmdLineArgs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr func(error) bool
	}{
		{
			name: "Endpoint missing",
			args: args{
				Arguments{
					ClusterID: "clusterid",
				},
				[]string{},
			},
			wantErr: errors.IsEndpointMissingError,
		},
		{
			name: "Auth token missing",
			args: args{
				Arguments{
					APIEndpoint: "https://some-endpoint.com",
					ClusterID: "clusterid",
				},
				[]string{},
			},
			wantErr: errors.IsNotLoggedInError,
		},
		{
			name: "Cluster ID missing",
			args: args{
				Arguments{
					APIEndpoint: "https://some-endpoint.com",
					AuthToken: "token",
					ClusterID: "",
				},
				[]string{},
			},
			wantErr: errors.IsClusterIDMissingError,
		},
	}
	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpgradeClusterPreconditions(tt.args.args, tt.args.cmdLineArgs)

			if err == nil {
				if tt.wantErr != nil {
					t.Errorf("Test case %d: Expected error, got nil", index)
				}
			} else {
				if !tt.wantErr(err) {
					t.Errorf("Test case %d: Didn't get the expected error type, got %s", index, err.Error())
				}
			}
		})
	}
}
