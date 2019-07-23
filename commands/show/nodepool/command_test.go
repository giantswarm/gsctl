package nodepool

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

func Test_ShowNodePool(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v5/clusters/cluster-id/nodepools/nodepool-id/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "nodepool-id",
					"name": "Application servers",
					"availability_zones": ["eu-west-1a", "eu-west-1c"],
					"scaling": {"min": 3, "max": 10},
					"node_spec": {"aws": {"instance_type": "c5.large"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
					"status": {"nodes": 3, "nodes_ready": 3}
				}`))
			default:
				t.Errorf("Unsupported route %s called in mock server", r.URL.Path)
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

	positionalArgs := []string{"cluster-id/nodepool-id"}

	flags.CmdAPIEndpoint = mockServer.URL
	flags.CmdToken = "my-token"
	args := defaultArguments(positionalArgs)

	err := verifyPreconditions(args, positionalArgs)
	if err != nil {
		t.Error(err)
	}

	result, err := fetchNodePool(args)
	if err != nil {
		t.Error(err)
	}

	if result.nodePool.ID != "nodepool-id" {
		t.Errorf("Got unexpected node pool ID %s", result.nodePool.ID)
	}

	if result.sumCPUs != 6 {
		t.Errorf("Got unexpected number of CPUs: %d", result.sumCPUs)
	}
	if result.sumMemory != 12.0 {
		t.Errorf("Got unexpected sum of RAM: %f GB", result.sumMemory)
	}

}
