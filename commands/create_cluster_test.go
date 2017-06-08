package commands

import (
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
		t.Error("Expected myDef.Workers[1].Memory.SizeGB to be 5, got: ", myDef.Workers[1].Memory.SizeGB)
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
		fmt.Printf("mockServer request: %s %s\n", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v4/clusters/f6e8r/")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"code": "RESOURCE_CREATED", "message": "Yeah!"}`))

	}))
	defer mockServer.Close()

	// first run, with minimal arguments
	args := []string{}
	cmdAPIEndpoint = mockServer.URL
	cmdOwner = "acme"
	checkAddCluster(CreateClusterCommand, args)
	addCluster(CreateClusterCommand, args)

	// second run, with additional arguments
	cmdClusterName = "UnitTestCluster"
	cmdNumWorkers = 4
	cmdKubernetesVersion = "myK8sVersion"
	cmdWorkerNumCPUs = 3
	cmdWorkerStorageSizeGB = 10
	cmdWorkerMemorySizeGB = 4
	checkAddCluster(CreateClusterCommand, args)
	addCluster(CreateClusterCommand, args)

	// third run, combining YAML and flags
	cmdInputYAMLFile = "testdata/minimal.yaml"
	cmdNumWorkers = 2
	cmdKubernetesVersion = "myK8sVersion"
	cmdWorkerNumCPUs = 4
	cmdWorkerStorageSizeGB = 20
	cmdWorkerMemorySizeGB = 6
	checkAddCluster(CreateClusterCommand, args)
	addCluster(CreateClusterCommand, args)

}
