package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giantswarm/microerror"
)

// TestReadFiles tests the readDefinitionFromFile with all
// YAML files in the testdata directory.
func TestReadFiles(t *testing.T) {
	basePath := "testdata"
	files, _ := ioutil.ReadDir(basePath)
	for _, f := range files {
		path := basePath + "/" + f.Name()
		_, err := readDefinitionFromFile(path)
		if err != nil {
			t.Error(err)
		}
	}
}

// Test_CreateFromYAML01 tests parsing a most simplistic YAML definition.
func Test_CreateFromYAML01(t *testing.T) {
	definition := clusterDefinition{}
	data := []byte(`owner: myorg`)
	myDef, err := unmarshalDefinition(data, definition)
	if err != nil {
		t.Error("Unmarshalling minimal cluster definition YAML failed: ", err)
	}
	if myDef.Owner != "myorg" {
		t.Error("Expected owner 'myorg', got: ", myDef.Owner)
	}
}

// Test_CreateFromYAML02 tests parsing a rather simplistic YAML definition.
func Test_CreateFromYAML02(t *testing.T) {
	definition := clusterDefinition{}
	data := []byte(`owner: myorg
name: Minimal cluster spec`)
	myDef, err := unmarshalDefinition(data, definition)
	if err != nil {
		t.Error("Unmarshalling minimal cluster definition YAML failed: ", err)
	}
	if myDef.Owner != "myorg" {
		t.Error("Expected owner 'myorg', got: ", myDef.Owner)
	}
	if myDef.Name != "Minimal cluster spec" {
		t.Error("Expected name 'Minimal cluster spec', got: ", myDef.Name)
	}
}

// Test_CreateFromYAML03 tests all the worker details.
func Test_CreateFromYAML03(t *testing.T) {
	definition := clusterDefinition{}
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
	myDef, err := unmarshalDefinition(data, definition)
	if err != nil {
		t.Error("Unmarshalling minimal cluster definition YAML failed: ", err)
	}
	if len(myDef.Workers) != 2 {
		t.Error("Expected 2 workers, got: ", len(myDef.Workers))
	}
	if myDef.Workers[1].CPU.Cores != 2 {
		t.Error("Expected myDef.Workers[1].CPU.Cores to be 2, got: ", myDef.Workers[1].CPU.Cores)
	}
	if myDef.Workers[1].Memory.SizeGB != 5.5 {
		t.Error("Expected myDef.Workers[1].Memory.SizeGB to be 5.5, got: ", myDef.Workers[1].Memory.SizeGB)
	}
	if myDef.Workers[1].Storage.SizeGB != 13.0 {
		t.Error("Expected myDef.Workers[1].Storage.SizeGB to be 13, got: ", myDef.Workers[1].Storage.SizeGB)
	}
}

// Test_CreateFromBadYAML01 tests how non-conforming YAML is treated.
func Test_CreateFromBadYAML01(t *testing.T) {
	definition := clusterDefinition{}
	data := []byte(`o: myorg`)
	myDef, err := unmarshalDefinition(data, definition)
	if err != nil {
		t.Error("Unmarshalling minimal cluster definition YAML failed: ", err)
	}
	if myDef.Owner != "" {
		t.Error("Expected owner to be empty, got: ", myDef.Owner)
	}
}

// Test_CreateClusterSuccessfully tests cluster creations that should succeed.
func Test_CreateClusterSuccessfully(t *testing.T) {
	var testCases = []struct {
		description string
		inputArgs   *addClusterArguments
	}{
		{
			description: "Minimal arguments",
			inputArgs: &addClusterArguments{
				owner: "acme",
				token: "fake token",
			},
		},
		{
			description: "Extensive arguments",
			inputArgs: &addClusterArguments{
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
			description: "Definition from YAML file",
			inputArgs: &addClusterArguments{
				clusterName:   "Cluster Name from Args",
				owner:         "acme",
				token:         "fake token",
				inputYAMLFile: "testdata/minimal.yaml",
				verbose:       true,
			},
		},
	}

	var validateErr error
	var executeErr error

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

		cmdAPIEndpoint = mockServer.URL
		cmdToken = testCase.inputArgs.token
		initClient()

		validateErr = validateCreateClusterPreConditions(*testCase.inputArgs)
		if validateErr != nil {
			t.Errorf("Validation error in testCase %d: %s", i, validateErr.Error())
		}
		_, executeErr = addCluster(*testCase.inputArgs)
		if executeErr != nil {
			t.Errorf("Execution error in testCase %d: %s", i, executeErr.Error())
		}
	}
}

// Test_CreateClusterExecutionFailures tests for errors thrown in the
// final execution of a cluster creations, which is the handling of the API call.
func Test_CreateClusterExecutionFailures(t *testing.T) {
	var testCases = []struct {
		description        string
		inputArgs          *addClusterArguments
		responseStatus     int
		serverResponseJSON []byte
		errorMatcher       func(err error) bool
	}{
		{
			description: "Unauthenticated request despite token being present",
			inputArgs: &addClusterArguments{
				owner: "owner",
				token: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "PERMISSION_DENIED", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusUnauthorized,
			errorMatcher:       IsNotAuthorizedError,
		},
		{
			description: "Owner organization not existing",
			inputArgs: &addClusterArguments{
				owner: "non-existing-owner",
				token: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusNotFound,
			errorMatcher:       IsOrganizationNotFoundError,
		},
		{
			description: "Non-existing YAML definition path",
			inputArgs: &addClusterArguments{
				owner:         "owner",
				token:         "some-token",
				inputYAMLFile: "does/not/exist.yaml",
				dryRun:        true,
			},
			serverResponseJSON: []byte(``),
			responseStatus:     0,
			errorMatcher:       IsYAMLFileNotReadableError,
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
		cmdAPIEndpoint = mockServer.URL // required to make initClient() work
		testCase.inputArgs.apiEndpoint = mockServer.URL
		err := initClient()
		if err != nil {
			t.Fatal(err)
		}

		validateErr := validateCreateClusterPreConditions(*testCase.inputArgs)
		if validateErr != nil {
			t.Errorf("Unexpected error in argument validation: %#v", validateErr)
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
