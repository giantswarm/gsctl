package commands

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestReadFiles tests the readDefinitionFromFile with all
// YAML files in the testdata directory
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

// Test_CreateFromYAML03 tests all the worker details
func Test_CreateFromYAML03(t *testing.T) {
	definition := clusterDefinition{}
	data := []byte(`owner: littleco
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

// Test_CreateFromBadYAML01 tests how non-conforming YAML is treated
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

// Test_CreateClusterSuccessfully tests cluster creations that should succeed
func Test_CreateClusterSuccessfully(t *testing.T) {
	// mock server always responding positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
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

	var testCases = []addClusterArguments{
		// minimal arguments
		{
			apiEndpoint: mockServer.URL,
			owner:       "acme",
			token:       "fake token",
		},
		// extensive arguments
		{
			apiEndpoint:         mockServer.URL,
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
		{
			apiEndpoint:   mockServer.URL,
			clusterName:   "Cluster Name from Args",
			owner:         "acme",
			token:         "fake token",
			inputYAMLFile: "testdata/minimal.yaml",
			verbose:       true,
		},
	}

	validateErr := errors.New("")
	executeErr := errors.New("")

	cmdAPIEndpoint = mockServer.URL
	initClient()

	for i, testCase := range testCases {
		validateErr = validateCreateClusterPreConditions(testCase)
		if validateErr != nil {
			t.Errorf("Validation error in testCase %v: %s", i, validateErr.Error())
		}
		_, executeErr = addCluster(testCase)
		if executeErr != nil {
			t.Errorf("Execution error in testCase %v: %s", i, executeErr.Error())
		}
	}
}

// Test_CreateClusterFailures tests for errors in cluster creations
// (these attempts should never succeed)
func Test_CreateClusterFailures(t *testing.T) {
	// mock server always responding negatively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Something was fishy"}`))
	}))
	defer mockServer.Close()

	var testCases = []addClusterArguments{
		// not authenticated
		{
			apiEndpoint: mockServer.URL,
			owner:       "owner",
			token:       "",
		},
		// extensive arguments (only a server error should let this fail)
		{
			apiEndpoint:         mockServer.URL,
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
		// file not readable
		{
			apiEndpoint:   mockServer.URL,
			token:         "fake token",
			inputYAMLFile: "does/not/exist.yaml",
			dryRun:        true,
		},
	}

	for i, testCase := range testCases {
		validateErr := validateCreateClusterPreConditions(testCase)
		if validateErr == nil {
			_, execErr := addCluster(testCase)
			if execErr == nil {
				t.Errorf("Expected errors didn't occur in testCase %v", i)
			}
		}
	}
}
