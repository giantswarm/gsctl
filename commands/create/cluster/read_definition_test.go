package cluster

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/commands/types"
)

// Test_ReadDefinitionFiles tests the readDefinitionFromFile with all
// YAML files in the testdata directory.
func Test_readDefinitionFromFile(t *testing.T) {
	basePath := "testdata"
	fs := afero.NewOsFs()

	var testCases = []struct {
		fileName     string
		errorMatcher func(error) bool
	}{
		{
			fileName:     "v4_complete_aws.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v4_complete.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v4_minimal.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v4_not_enough_workers.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v5_minimal.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v5_three_nodepools.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v5_three_nodepools_azure.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v5_with_ha_master.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "v5_instance_distribution.yaml",
			errorMatcher: nil,
		},
		{
			fileName:     "invalid01.yaml",
			errorMatcher: IsInvalidDefinitionYAML,
		},
		{
			fileName:     "invalid02.yaml",
			errorMatcher: IsInvalidDefinitionYAML,
		},
		{
			fileName:     "invalid03.yaml",
			errorMatcher: IsInvalidDefinitionYAML,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d, file %s", i, tc.fileName)
		path := basePath + "/" + tc.fileName

		_, err := readDefinitionFromFile(fs, path)
		if tc.errorMatcher != nil {
			if !tc.errorMatcher(err) {
				t.Errorf("Unexpected error in case %d, file %s: %s", i, tc.fileName, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error in case %d, file %s: %s", i, tc.fileName, err)
		}
	}
}

// Test_ParseYAMLDefinitionV4 tests parsing v4 YAML definition files.
func Test_ParseYAMLDefinitionV4(t *testing.T) {
	var testCases = []struct {
		inputYAML      []byte
		expectedOutput *types.ClusterDefinitionV4
		errorMatcher   func(error) bool
	}{
		// Minimal YAML.
		{
			inputYAML: []byte(`owner: myorg`),
			expectedOutput: &types.ClusterDefinitionV4{
				Owner: "myorg",
			},
		},
		// Invalid YAML.
		{
			inputYAML:      []byte(`owner\n    foo`),
			expectedOutput: nil,
			errorMatcher:   IsUnmashalToMapFailed,
		},
		// More details.
		{
			inputYAML: []byte(`owner: myorg
name: My cluster
release_version: 1.2.3
availability_zones: 3
scaling:
  min: 3
  max: 5`),
			expectedOutput: &types.ClusterDefinitionV4{
				Owner:             "myorg",
				Name:              "My cluster",
				ReleaseVersion:    "1.2.3",
				AvailabilityZones: 3,
				Scaling: types.ScalingDefinition{
					Min: 3,
					Max: 5,
				},
			},
		},
		// KVM worker details.
		{
			inputYAML: []byte(`owner: myorg
workers:
- memory:
    size_gb: 16.5
  cpu:
    cores: 4
  storage:
    size_gb: 100
- memory:
    size_gb: 32
  cpu:
    cores: 8
  storage:
    size_gb: 50
`),
			expectedOutput: &types.ClusterDefinitionV4{
				Owner: "myorg",
				Workers: []types.NodeDefinition{
					{
						Memory:  types.MemoryDefinition{SizeGB: 16.5},
						CPU:     types.CPUDefinition{Cores: 4},
						Storage: types.StorageDefinition{SizeGB: 100},
					},
					{
						Memory:  types.MemoryDefinition{SizeGB: 32},
						CPU:     types.CPUDefinition{Cores: 8},
						Storage: types.StorageDefinition{SizeGB: 50},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			def, err := readDefinitionFromYAML(tc.inputYAML)
			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Expected error, got %v", i, err)
				}
			} else if err != nil {
				t.Errorf("Case %d - Unexpected error %v", i, err)
			}

			if err == nil {
				if diff := cmp.Diff(tc.expectedOutput, def); diff != "" {
					t.Errorf("Case %d - Resulting definition unequal. (-expected +got):\n%s", i, diff)
				}
			}
		})
	}
}

// Test_ParseYAMLDefinitionV5 tests parsing v5 YAML definition files.
func Test_ParseYAMLDefinitionV5(t *testing.T) {
	var testCases = []struct {
		inputYAML      []byte
		expectedOutput *types.ClusterDefinitionV5
	}{
		// Minimal YAML.
		{
			[]byte(`api_version: v5
owner: myorg`),
			&types.ClusterDefinitionV5{
				APIVersion: "v5",
				Owner:      "myorg",
			},
		},
		// More details.
		{
			[]byte(`api_version: v5
owner: myorg
name: My cluster
release_version: 1.2.3
`),
			&types.ClusterDefinitionV5{
				APIVersion:     "v5",
				Owner:          "myorg",
				Name:           "My cluster",
				ReleaseVersion: "1.2.3",
			},
		},
		// Node pools.
		{
			[]byte(`api_version: v5
owner: myorg
master:
  availability_zone: my-zone-1a
nodepools:
- name: General purpose
  availability_zones:
    number: 2
- name: Database
  availability_zones:
    zones:
    - my-zone-1a
    - my-zone-1b
    - my-zone-1c
  scaling:
    min: 3
    max: 10
  node_spec:
    aws:
      instance_type: "m5.superlarge"
- name: Batch
`),
			&types.ClusterDefinitionV5{
				APIVersion: "v5",
				Owner:      "myorg",
				Master:     &types.MasterDefinition{AvailabilityZone: "my-zone-1a"},
				NodePools: []*types.NodePoolDefinition{
					{
						Name:              "General purpose",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Number: 2},
					},
					{
						Name:              "Database",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Zones: []string{"my-zone-1a", "my-zone-1b", "my-zone-1c"}},
						Scaling:           &types.ScalingDefinition{Min: 3, Max: 10},
						NodeSpec:          &types.NodeSpec{AWS: &types.AWSSpecificDefinition{InstanceType: "m5.superlarge"}},
					},
					{
						Name: "Batch",
					},
				},
			},
		},
		// HA master.
		{
			[]byte(`api_version: v5
owner: myorg
master_nodes:
  high_availability: true
nodepools:
- name: General purpose
  availability_zones:
    number: 2
- name: Database
  availability_zones:
    zones:
    - my-zone-1a
    - my-zone-1b
    - my-zone-1c
  scaling:
    min: 3
    max: 10
  node_spec:
    aws:
      instance_type: "m5.superlarge"
- name: Batch
`),
			&types.ClusterDefinitionV5{
				APIVersion:  "v5",
				Owner:       "myorg",
				MasterNodes: &types.MasterNodes{HighAvailability: toBoolPtr(true)},
				NodePools: []*types.NodePoolDefinition{
					{
						Name:              "General purpose",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Number: 2},
					},
					{
						Name:              "Database",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Zones: []string{"my-zone-1a", "my-zone-1b", "my-zone-1c"}},
						Scaling:           &types.ScalingDefinition{Min: 3, Max: 10},
						NodeSpec:          &types.NodeSpec{AWS: &types.AWSSpecificDefinition{InstanceType: "m5.superlarge"}},
					},
					{
						Name: "Batch",
					},
				},
			},
		},
		{
			// like testdata/v5_instance_distribution.yaml
			[]byte(`api_version: v5
release_version: "11.5.0"
owner: acme
name: Cluster with several node pools testing various instance distribution combinations
nodepools:
- name: Node pool with 0 on-demand, 100% spot, no alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 0
        on_demand_percentage_above_base_capacity: 0
      use_alike_instance_types: false
- name: Node pool with 3 on-demand, 100% spot, no alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 3
        on_demand_percentage_above_base_capacity: 0
      use_alike_instance_types: false
- name: Node pool with 3 on-demand, 50% spot, no alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 3
        on_demand_percentage_above_base_capacity: 50
      use_alike_instance_types: false
- name: Node pool with 0 on-demand, 100% spot, use alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 0
        on_demand_percentage_above_base_capacity: 0
      use_alike_instance_types: true
- name: Node pool with 3 on-demand, 100% spot, use alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 3
        on_demand_percentage_above_base_capacity: 0
      use_alike_instance_types: true
- name: Node pool with 3 on-demand, 50% spot, use alike instance types
  node_spec:
    aws:
      instance_distribution:
        on_demand_base_capacity: 3
        on_demand_percentage_above_base_capacity: 50
      use_alike_instance_types: true
`),
			&types.ClusterDefinitionV5{
				APIVersion:     "v5",
				ReleaseVersion: "11.5.0",
				Name:           "Cluster with several node pools testing various instance distribution combinations",
				Owner:          "acme",
				Master:         nil,
				MasterNodes:    nil,
				NodePools: []*types.NodePoolDefinition{
					{
						Name: "Node pool with 0 on-demand, 100% spot, no alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                0,
									OnDemandPercentageAboveBaseCapacity: 0,
								},
								UseAlikeInstanceTypes: false,
							},
						},
					},
					{
						Name: "Node pool with 3 on-demand, 100% spot, no alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                3,
									OnDemandPercentageAboveBaseCapacity: 0,
								},
								UseAlikeInstanceTypes: false,
							},
						},
					},
					{
						Name: "Node pool with 3 on-demand, 50% spot, no alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                3,
									OnDemandPercentageAboveBaseCapacity: 50,
								},
								UseAlikeInstanceTypes: false,
							},
						},
					},
					{
						Name: "Node pool with 0 on-demand, 100% spot, use alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                0,
									OnDemandPercentageAboveBaseCapacity: 0,
								},
								UseAlikeInstanceTypes: true,
							},
						},
					},
					{
						Name: "Node pool with 3 on-demand, 100% spot, use alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                3,
									OnDemandPercentageAboveBaseCapacity: 0,
								},
								UseAlikeInstanceTypes: true,
							},
						},
					},
					{
						Name: "Node pool with 3 on-demand, 50% spot, use alike instance types",
						NodeSpec: &types.NodeSpec{
							AWS: &types.AWSSpecificDefinition{
								InstanceDistribution: &types.AWSInstanceDistribution{
									OnDemandBaseCapacity:                3,
									OnDemandPercentageAboveBaseCapacity: 50,
								},
								UseAlikeInstanceTypes: true,
							},
						},
					},
				},
			},
		},
		// Azure Node pools.
		{
			[]byte(`api_version: v5
owner: myorg
master:
  availability_zone: 2
nodepools:
- name: General purpose
  availability_zones:
    number: 2
- name: Database
  availability_zones:
    zones:
    - 1
    - 2
    - 3
  node_spec:
    azure:
      vm_size: "Standard_D4s_v3"
- name: Batch
`),
			&types.ClusterDefinitionV5{
				APIVersion: "v5",
				Owner:      "myorg",
				Master:     &types.MasterDefinition{AvailabilityZone: "2"},
				NodePools: []*types.NodePoolDefinition{
					{
						Name:              "General purpose",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Number: 2},
					},
					{
						Name:              "Database",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Zones: []string{"1", "2", "3"}},
						NodeSpec:          &types.NodeSpec{Azure: &types.AzureSpecificDefinition{VMSize: "Standard_D4s_v3"}},
					},
					{
						Name: "Batch",
					},
				},
			},
		},
		// Azure spot instances.
		{
			[]byte(`api_version: v5
owner: myorg
nodepools:
- name: Database
  availability_zones:
    zones:
    - 1
    - 2
    - 3
  node_spec:
    azure:
      spot_instances:
        enabled: true
        max_price: 0.01235
`),
			&types.ClusterDefinitionV5{
				APIVersion: "v5",
				Owner:      "myorg",
				NodePools: []*types.NodePoolDefinition{
					{
						Name:              "Database",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Zones: []string{"1", "2", "3"}},
						NodeSpec: &types.NodeSpec{Azure: &types.AzureSpecificDefinition{
							AzureSpotInstances: &types.AzureSpotInstances{
								Enabled:  true,
								MaxPrice: 0.01235,
							},
						},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			def, err := readDefinitionFromYAML(tc.inputYAML)
			if err != nil {
				t.Errorf("Case %d - Unexpected error %v", i, err)
			}

			if diff := cmp.Diff(tc.expectedOutput, def); diff != "" {
				t.Errorf("Case %d - Resulting definition unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// Test_CreateFromBadYAML01 tests how non-conforming YAML is treated.
func Test_CreateFromBadYAML01(t *testing.T) {
	data := []byte(`o: myorg`)
	def := types.ClusterDefinitionV4{}

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if def.Owner != "" {
		t.Fatalf("expected owner to be empty, got %q", def.Owner)
	}
}
