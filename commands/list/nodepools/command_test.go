package nodepools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

func Test_ListNodePools(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v5/clusters/cluster-id/nodepools/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{"id": "a7rc4", "name": "Batch number crunching", "availability_zones": ["eu-west-1d"], "scaling": {"min": 2, "max": 5}, "node_spec": {"aws": {"instance_type": "p3.8xlarge"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 4, "nodes_ready": 4}},
					{"id": "6feel", "name": "Application servers", "availability_zones": ["eu-west-1a", "eu-west-1b", "eu-west-1c"], "scaling": {"min": 3, "max": 15}, "node_spec": {"aws": {"instance_type": "p3.2xlarge"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 10, "nodes_ready": 9}},
					{"id": "a6bf4", "name": "New node pool", "availability_zones": ["eu-west-1c"], "scaling": {"min": 3, "max": 3}, "node_spec": {"aws": {"instance_type": "m5.2xlarge"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
				]`))
			default:
				t.Errorf("Unsupported route %scalled in mock server", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
			}
		}
	}))
	defer mockServer.Close()

	// temp config
	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	positionalArgs := []string{"cluster-id"}

	flags.CmdAPIEndpoint = mockServer.URL
	flags.CmdToken = "my-token"
	args := defaultArgs(positionalArgs)

	err := verifyPreconditions(args, positionalArgs)
	if err != nil {
		t.Error(err)
	}

	results, err := fetchNodePools(args)
	if err != nil {
		t.Error(err)
	}

	if len(results) != 3 {
		t.Errorf("Expected length 3, got %d", len(results))
	}

	if results[0].nodePool.ID != "6feel" {
		t.Errorf("Expected NP ID 6feel in the front, got %s", results[0].nodePool.ID)
	}

	if results[2].sumCPUs != 128 {
		t.Errorf("Expected 128 CPUs in the third row, got %d", results[2].sumCPUs)
	}
}
