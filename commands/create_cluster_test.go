package commands

import (
	"errors"
	"fmt"
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

// Test_CreateFromCommandLine tests a cluster creation completely based on
// command line arguments
func Test_CreateFromCommandLine(t *testing.T) {
	// mock server always responding positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v4/clusters/f6e8r/")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"code": "RESOURCE_CREATED", "message": "Yeah!"}`))
	}))
	defer mockServer.Close()

	var testCases = []addClusterArguments{
		// minimal arguments
		addClusterArguments{
			apiEndpoint: mockServer.URL,
			owner:       "acme",
			token:       "fake token",
		},
		// extensive arguments
		addClusterArguments{
			apiEndpoint:         mockServer.URL,
			clusterName:         "UnitTestCluster",
			numWorkers:          4,
			kubernetesVersion:   "myK8sVersion",
			owner:               "acme",
			token:               "fake token",
			workerNumCPUs:       3,
			workerMemorySizeGB:  4,
			workerStorageSizeGB: 10,
			verbose:             true,
		},
		addClusterArguments{
			apiEndpoint:       mockServer.URL,
			clusterName:       "Cluster Name from Args",
			owner:             "acme",
			token:             "fake token",
			inputYAMLFile:     "testdata/minimal.yaml",
			kubernetesVersion: "K8sVersionFromArgs",
			verbose:           true,
		},
	}

	validateErr := errors.New("")
	executeErr := errors.New("")
	//result := addClusterResult{}

	for i, testCase := range testCases {
		validateErr = validatePreConditions(testCase)
		if validateErr != nil {
			t.Error(fmt.Sprintf("Validation error in testCase %v: %s", i, validateErr.Error()))
		}
		_, executeErr = addCluster(testCase)
		if executeErr != nil {
			t.Error(fmt.Sprintf("Execution error in testCase %v: %s", i, executeErr.Error()))
		}
	}

}
