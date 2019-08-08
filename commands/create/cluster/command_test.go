package cluster

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/flags"
)

// TestReadFiles tests the readDefinitionFromFile with all
// YAML files in the testdata directory.
func TestReadFiles(t *testing.T) {
	basePath := "testdata"
	fs := afero.NewOsFs()
	files, _ := afero.ReadDir(fs, basePath)
	for _, f := range files {
		path := basePath + "/" + f.Name()
		_, err := readDefinitionFromFile(fs, path)
		if err != nil {
			t.Error(err)
		}
	}
}

// Test_CreateFromYAML01 tests parsing a most simplistic YAML definition.
func Test_CreateFromYAML01(t *testing.T) {
	def := types.ClusterDefinition{}
	data := []byte(`owner: myorg`)

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if def.Owner != "myorg" {
		t.Error("expected owner 'myorg', got: ", def.Owner)
	}
}

// Test_CreateFromYAML02 tests parsing a rather simplistic YAML definition.
func Test_CreateFromYAML02(t *testing.T) {
	def := types.ClusterDefinition{}
	data := []byte(`
owner: myorg
name: Minimal cluster spec
`)

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if def.Owner != "myorg" {
		t.Error("expected owner 'myorg', got: ", def.Owner)
	}
	if def.Name != "Minimal cluster spec" {
		t.Error("expected name 'Minimal cluster spec', got: ", def.Name)
	}
}

// Test_CreateFromYAML03 tests all the worker details.
func Test_CreateFromYAML03(t *testing.T) {
	def := types.ClusterDefinition{}
	data := []byte(`
owner: littleco
workers:
  - memory:
    size_gb: 2
  - cpu:
      cores: 2
    memory:
      size_gb: 5.5
    storage:
      size_gb: 13
    labels:
      foo: bar
`)

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if len(def.Workers) != 2 {
		t.Error("expected 2 workers, got: ", len(def.Workers))
	}
	if def.Workers[1].CPU.Cores != 2 {
		t.Error("expected def.Workers[1].CPU.Cores to be 2, got: ", def.Workers[1].CPU.Cores)
	}
	if def.Workers[1].Memory.SizeGB != 5.5 {
		t.Error("expected def.Workers[1].Memory.SizeGB to be 5.5, got: ", def.Workers[1].Memory.SizeGB)
	}
	if def.Workers[1].Storage.SizeGB != 13.0 {
		t.Error("expected def.Workers[1].Storage.SizeGB to be 13, got: ", def.Workers[1].Storage.SizeGB)
	}
}

// Test_CreateFromBadYAML01 tests how non-conforming YAML is treated.
func Test_CreateFromBadYAML01(t *testing.T) {
	data := []byte(`o: myorg`)
	def := types.ClusterDefinition{}

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if def.Owner != "" {
		t.Fatalf("expected owner to be empty, got %q", def.Owner)
	}
}

// Test_CreateClusterSuccessfully tests cluster creations that should succeed.
func Test_CreateClusterSuccessfully(t *testing.T) {
	var testCases = []struct {
		description string
		inputArgs   *Arguments
	}{
		{
			description: "Minimal arguments",
			inputArgs: &Arguments{
				owner: "acme",
				token: "fake token",
			},
		},
		{
			description: "Extensive arguments",
			inputArgs: &Arguments{
				clusterName:         "UnitTestCluster",
				numWorkers:          4,
				releaseVersion:      "0.3.0",
				owner:               "acme",
				token:               "fake token",
				workerNumCPUs:       3,
				workerMemorySizeGB:  4,
				workerStorageSizeGB: 10,
				verbose:             true,
			},
		},
		{
			description: "Max workers",
			inputArgs: &Arguments{
				owner:      "acme",
				workersMax: 4,
				token:      "fake token",
			},
		},
		{
			description: "Min workers",
			inputArgs: &Arguments{
				owner:      "acme",
				workersMin: 4,
				token:      "fake token",
			},
		},
		{
			description: "Min workers and max workers same",
			inputArgs: &Arguments{
				owner:      "acme",
				workersMin: 4,
				workersMax: 4,
				token:      "fake token",
			},
		},
		{
			description: "Min workers and max workers different",
			inputArgs: &Arguments{
				owner:      "acme",
				workersMin: 2,
				workersMax: 4,
				token:      "fake token",
			},
		},
		{
			description: "Definition from YAML file",
			inputArgs: &Arguments{
				clusterName:   "Cluster Name from Args",
				fileSystem:    afero.NewOsFs(), // needed for YAML file access
				inputYAMLFile: "testdata/minimal.yaml",
				owner:         "acme",
				token:         "fake token",
				verbose:       true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Logf("Case %d: %s", i, testCase.description)

		// mock server always responding positively
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("mockServer request: ", r.Method, r.URL)
			w.Header().Set("Content-Type", "application/json")
			if !strings.Contains(r.Header.Get("Authorization"), testCase.inputArgs.token) {
				t.Errorf("Authorization header incomplete: '%s'", r.Header.Get("Authorization"))
			}
			if r.Method == "POST" && r.URL.String() == "/v4/clusters/" {
				w.Header().Set("Location", "/v4/clusters/f6e8r/")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"code": "RESOURCE_CREATED", "message": "Yeah!"}`))
			} else if r.Method == "GET" && r.URL.String() == "/v4/releases/" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
			  {
					"timestamp": "2017-10-15T12:00:00Z",
			    "version": "0.3.0",
			    "active": true,
			    "changelog": [
			      {
			        "component": "firstComponent",
			        "description": "firstComponent added."
			      }
			    ],
			    "components": [
			      {
			        "name": "firstComponent",
			        "version": "0.0.1"
			      }
			    ]
			  }
			]`))
			}
		}))
		defer mockServer.Close()

		testCase.inputArgs.apiEndpoint = mockServer.URL
		testCase.inputArgs.userProvidedToken = testCase.inputArgs.token

		err := validatePreConditions(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Validation error in testCase %d: %s", i, err.Error())
		}
		_, err = addCluster(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Execution error in testCase %d: %s", i, err.Error())
		}
	}
}

// Test_CreateClusterExecutionFailures tests for errors thrown in the
// final execution of a cluster creations, which is the handling of the API call.
func Test_CreateClusterExecutionFailures(t *testing.T) {
	var testCases = []struct {
		description        string
		inputArgs          *Arguments
		responseStatus     int
		serverResponseJSON []byte
		errorMatcher       func(err error) bool
	}{
		{
			description: "Unauthenticated request despite token being present",
			inputArgs: &Arguments{
				owner: "owner",
				token: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "PERMISSION_DENIED", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusUnauthorized,
			errorMatcher:       errors.IsNotAuthorizedError,
		},
		{
			description: "Owner organization not existing",
			inputArgs: &Arguments{
				owner: "non-existing-owner",
				token: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusNotFound,
			errorMatcher:       errors.IsOrganizationNotFoundError,
		},
		{
			description: "Non-existing YAML definition path",
			inputArgs: &Arguments{
				owner:         "owner",
				token:         "some-token",
				fileSystem:    afero.NewOsFs(),
				inputYAMLFile: "does/not/exist.yaml",
				dryRun:        true,
			},
			serverResponseJSON: []byte(``),
			responseStatus:     0,
			errorMatcher:       errors.IsYAMLFileNotReadableError,
		},
	}

	for i, testCase := range testCases {
		t.Logf("Case %d: %s", i, testCase.description)

		// mock server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//t.Log("mockServer request: ", r.Method, r.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(testCase.responseStatus)
			w.Write([]byte(testCase.serverResponseJSON))
		}))
		defer mockServer.Close()

		// client
		flags.CmdAPIEndpoint = mockServer.URL // required to make InitClient() work
		testCase.inputArgs.apiEndpoint = mockServer.URL

		err := validatePreConditions(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Unexpected error in argument validation: %#v", err)
		} else {
			_, err := addCluster(*testCase.inputArgs)
			if err == nil {
				t.Errorf("Test case %d did not yield an execution error.", i)
			}
			origErr := microerror.Cause(err)
			if testCase.errorMatcher(origErr) == false {
				t.Errorf("Test case %d did not yield the expected execution error, instead: %#v", i, err)
			}
		}
	}
}

func Test_CreateCluster_ValidationFailures(t *testing.T) {
	var testCases = []struct {
		name         string
		inputArgs    *Arguments
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0 workers min is higher than max",
			inputArgs: &Arguments{
				owner:      "owner",
				token:      "some-token",
				workersMin: 4,
				workersMax: 2,
			},
			errorMatcher: errors.IsWorkersMinMaxInvalid,
		},
		{
			name: "case 1 workers min and max with legacy num workers",
			inputArgs: &Arguments{
				owner:      "owner",
				token:      "some-token",
				workersMin: 4,
				workersMax: 2,
				numWorkers: 2,
			},
			errorMatcher: errors.IsConflictingWorkerFlagsUsed,
		},
		{
			name: "case 2 workers min with legacy num workers",
			inputArgs: &Arguments{
				owner:      "owner",
				token:      "some-token",
				workersMin: 4,
				numWorkers: 2,
			},
			errorMatcher: errors.IsConflictingWorkerFlagsUsed,
		},
		{
			name: "case 3 workers max with legacy num workers",
			inputArgs: &Arguments{
				owner:      "owner",
				token:      "some-token",
				workersMax: 2,
				numWorkers: 2,
			},
			errorMatcher: errors.IsConflictingWorkerFlagsUsed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePreConditions(*tc.inputArgs)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}
		})
	}
}
