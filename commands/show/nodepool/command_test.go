package nodepool

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

// Test_ShowNodePool tries cases where we don't expect any errors.
func Test_ShowNodePool(t *testing.T) {
	var testCases = []struct {
		responseBody string
		sumCPUs      int64
		sumMemory    float64
	}{
		{
			`{
				"id": "nodepool-id",
				"name": "Application servers",
				"availability_zones": ["eu-west-1a", "eu-west-1c"],
				"scaling": {"Min": 3, "Max": 10},
				"node_spec": {"aws": {"instance_type": "c5.large"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 3, "nodes_ready": 3},
				"subnet": "10.1.0.0/24"
			}`,
			6,
			12.0,
		},
		{
			// Instance type "nonexisting" does not exist. That's on purpose.
			`{
				"id":"nodepool-id",
				"name":"awesome-nodepool",
				"availability_zones":["europe-west-1b","europe-central-1a","europe-central-1b"],
				"scaling":{"min":2,"max":5},
				"node_spec":{"aws":{"instance_type":"nonexisting"},"labels":["web-compute"],"volume_sizes":{"docker":100,"kubelet":100}},
				"status":{"nodes":4,"nodes_ready":4},
				"subnet":"10.1.0.0/24"
			}`,
			0,
			0.0,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "GET" {
					switch uri := r.URL.Path; uri {
					case "/v5/clusters/cluster-id/nodepools/nodepool-id/":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tc.responseBody))
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
			flags.CmdVerbose = true
			args := defaultArguments(positionalArgs)

			err := verifyPreconditions(args, positionalArgs)
			if err != nil {
				t.Errorf("Case %d: unexpected error %s", i, err)
			}

			result, err := fetchNodePool(args)
			if err != nil {
				t.Errorf("Case %d: unexpected error '%s'", i, err)
			}

			if result == nil {
				t.Fatalf("Case %d: Got Got nil instead of node pool details", i)
			}

			if result.nodePool.ID != "nodepool-id" {
				t.Errorf("Case %d: Got unexpected node pool ID %s", i, result.nodePool.ID)
			}

			if result.sumCPUs != tc.sumCPUs {
				t.Errorf("Case %d: Got unexpected number of CPUs: %d", i, result.sumCPUs)
			}
			if result.sumMemory != tc.sumMemory {
				t.Errorf("Case %d: Got unexpected sum of RAM: %f GB", i, result.sumMemory)
			}
		})
	}

}
