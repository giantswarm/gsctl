package nodepools

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/gscliauth/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/testutils"
)

func Test_ListNodePools(t *testing.T) {
	testCases := []struct {
		npResponse   string
		outputFormat string
		output       string
	}{
		{
			npResponse: `[
                {"id": "a7rc4", "name": "Batch number crunching", "availability_zones": ["eu-west-1d"], "scaling": {"min": 2, "max": 5}, "node_spec": {"aws": {"instance_type": "p3.8xlarge", "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 4, "nodes_ready": 4}},
                {"id": "6feel", "name": "Application servers", "availability_zones": ["eu-west-1a", "eu-west-1b", "eu-west-1c"], "scaling": {"min": 3, "max": 15}, "node_spec": {"aws": {"instance_type": "p3.2xlarge", "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 10, "nodes_ready": 9}},
                {"id": "a6bf4", "name": "New node pool", "availability_zones": ["eu-west-1c"], "scaling": {"min": 3, "max": 3}, "node_spec": {"aws": {"instance_type": "m5.2xlarge", "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "table",
			output: `ID     NAME                    AZ     INSTANCE TYPE  ALIKE  ON-DEMAND BASE  SPOT PERCENTAGE  NODES MIN/MAX  NODES DESIRED  NODES READY  SPOT INSTANCES  CPUS  RAM (GB)
6feel  Application servers     A,B,C  p3.2xlarge     false               0              100           3/15             10            9               0    72     549.0
a6bf4  New node pool           C      m5.2xlarge     false               0              100            3/3              0            0               0     0       0.0
a7rc4  Batch number crunching  D      p3.8xlarge     false               0              100            2/5              4            4               0   128     976.0`,
		},
		{
			npResponse: `[
                {"id": "a6bf4", "name": "New node pool", "availability_zones": ["eu-west-1c"], "scaling": {"min": 3, "max": 3}, "node_spec": {"aws": {"instance_type": "m5.2xlarge", "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "table",
			output: `ID     NAME           AZ  INSTANCE TYPE  ALIKE  ON-DEMAND BASE  SPOT PERCENTAGE  NODES MIN/MAX  NODES DESIRED  NODES READY  SPOT INSTANCES  CPUS  RAM (GB)
a6bf4  New node pool  C   m5.2xlarge     false               0              100            3/3              0            0               0     0       0.0`,
		},
		{
			npResponse: `[
                {"id": "a6bf4", "name": "New node pool", "availability_zones": ["eu-west-1c"], "scaling": {"min": 3, "max": 3}, "node_spec": {"aws": {"instance_type": "m5.2xlarge", "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "json",
			output: `[
  {
    "availability_zones": [
      "eu-west-1c"
    ],
    "id": "a6bf4",
    "name": "New node pool",
    "node_spec": {
      "aws": {
        "instance_distribution": {},
        "instance_type": "m5.2xlarge"
      },
      "volume_sizes_gb": {
        "docker": 100,
        "kubelet": 100
      }
    },
    "scaling": {
      "max": 3,
      "min": 3
    },
    "status": {
      "instance_types": null
    }
  }
]`,
		},
		{
			npResponse: `[
                {"id": "np001-1", "name": "np001-1", "availability_zones": ["2"], "scaling": {"min": 1, "max": 2}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "table",
			output: `ID       NAME     AZ  VM SIZE          NODES MIN/MAX  NODES DESIRED  NODES READY  SPOT INSTANCES  CPUS  RAM (GB)
np001-1  np001-1  2   Standard_D4s_v3            1/2              0            0             OFF  0     0.0`,
		},
		{
			npResponse: `[
                {"id": "np002-1", "name": "np002-1", "availability_zones": ["1"], "scaling": {"min": 1, "max": 2}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}},
                {"id": "np002-2", "name": "np002-2", "availability_zones": ["2", "3"], "scaling": {"min": 1, "max": 3}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3", "spot_instances": {"enabled": true, "max_price": 0.05312}}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}},
                {"id": "np002-3", "name": "np002-3", "availability_zones": ["2"], "scaling": {"min": 3, "max": 10}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3", "spot_instances": {"enabled": false}}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "table",
			output: `ID       NAME     AZ   VM SIZE          NODES MIN/MAX  NODES DESIRED  NODES READY  SPOT INSTANCES  CPUS  RAM (GB)
np002-1  np002-1  1    Standard_D4s_v3            1/2              0            0             OFF  0     0.0
np002-2  np002-2  2,3  Standard_D4s_v3            1/3              0            0              ON  0     0.0
np002-3  np002-3  2    Standard_D4s_v3           3/10              0            0             OFF  0     0.0`,
		},
		{
			npResponse: `[
                {"id": "np001-1", "name": "np001-1", "availability_zones": ["2"], "scaling": {"min": -1, "max": -1}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "json",
			output: `[
  {
    "availability_zones": [
      "2"
    ],
    "id": "np001-1",
    "name": "np001-1",
    "node_spec": {
      "azure": {
        "vm_size": "Standard_D4s_v3"
      },
      "volume_sizes_gb": {
        "docker": 50,
        "kubelet": 100
      }
    },
    "scaling": {
      "max": -1,
      "min": -1
    },
    "status": {
      "instance_types": null
    }
  }
]`,
		},
		{
			npResponse: `[
                {"id": "np002-1", "name": "np002-1", "availability_zones": ["1"], "scaling": {"min": -1, "max": -1}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}},
                {"id": "np002-2", "name": "np002-2", "availability_zones": ["2", "3"], "scaling": {"min": -1, "max": -1}, "node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}, "volume_sizes_gb": {"docker": 50, "kubelet": 100}}, "status": {"nodes": 0, "nodes_ready": 0}}
            ]`,
			outputFormat: "json",
			output: `[
  {
    "availability_zones": [
      "1"
    ],
    "id": "np002-1",
    "name": "np002-1",
    "node_spec": {
      "azure": {
        "vm_size": "Standard_D4s_v3"
      },
      "volume_sizes_gb": {
        "docker": 50,
        "kubelet": 100
      }
    },
    "scaling": {
      "max": -1,
      "min": -1
    },
    "status": {
      "instance_types": null
    }
  },
  {
    "availability_zones": [
      "2",
      "3"
    ],
    "id": "np002-2",
    "name": "np002-2",
    "node_spec": {
      "azure": {
        "vm_size": "Standard_D4s_v3"
      },
      "volume_sizes_gb": {
        "docker": 50,
        "kubelet": 100
      }
    },
    "scaling": {
      "max": -1,
      "min": -1
    },
    "status": {
      "instance_types": null
    }
  }
]`,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "GET" {
					switch uri := r.URL.Path; uri {
					case "/v5/clusters/cluster-id/nodepools/":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tc.npResponse))

					case "/v4/clusters/":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`[
					{
						"id": "cluster-id",
						"name": "Name of the cluster",
						"owner": "acme"
					}
				]`))

					default:
						t.Errorf("Case %d: Unsupported route %s called in mock server", i, r.URL.Path)
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

			args := Arguments{
				clusterNameOrID: "cluster-id",
				apiEndpoint:     mockServer.URL,
				authToken:       "my-token",
				outputFormat:    tc.outputFormat,
			}

			err := verifyPreconditions(args, []string{args.clusterNameOrID})
			if err != nil {
				t.Errorf("Case %d: %s", i, err)
			}

			results, err := fetchNodePools(args)
			if err != nil {
				t.Errorf("Case %d: %s", i, err)
			}

			output, err := getOutput(results, args.outputFormat)
			if err != nil {
				t.Errorf("Case %d: %s", i, err)
			}

			if diff := cmp.Diff(tc.output, output); diff != "" {
				t.Errorf("Case %d - Command output is incorrect. (-expected +got):\n%s", i, diff)
			}
		})
	}
}
