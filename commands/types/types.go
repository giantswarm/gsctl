// Package types defines types to be used in several commands.
package types

// CPUDefinition defines worker node CPU specs.
type CPUDefinition struct {
	Cores int `yaml:"cores,omitempty"`
}

// MemoryDefinition defines worker node memory specs.
type MemoryDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

// StorageDefinition defines worker node storage specs.
type StorageDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

// AWSSpecificDefinition defines worker node specs for AWS.
type AWSSpecificDefinition struct {
	InstanceDistribution  *AWSInstanceDistribution `yaml:"instance_distribution,omitempty"`
	InstanceType          string                   `yaml:"instance_type,omitempty"`
	UseAlikeInstanceTypes bool                     `yaml:"use_alike_instance_types,omitempty"`
}

// AWSInstanceDistribution defines the distribution between on-demand and spot instances.
type AWSInstanceDistribution struct {
	OnDemandBaseCapacity                int64 `yaml:"on_demand_base_capacity"`
	OnDemandPercentageAboveBaseCapacity int64 `yaml:"on_demand_percentage_above_base_capacity"`
}

// AzureSpecificDefinition defines worker node specs for Azure.
type AzureSpecificDefinition struct {
	VMSize string `yaml:"vm_size,omitempty"`
}

// NodeDefinition defines worker node specs.
type NodeDefinition struct {
	Memory  MemoryDefinition        `yaml:"memory,omitempty"`
	CPU     CPUDefinition           `yaml:"cpu,omitempty"`
	Storage StorageDefinition       `yaml:"storage,omitempty"`
	Labels  map[string]string       `yaml:"labels,omitempty"`
	AWS     AWSSpecificDefinition   `yaml:"aws,omitempty"`
	Azure   AzureSpecificDefinition `yaml:"azure,omitempty"`
}

// ClusterDefinitionV4 defines a tenant cluster spec compatible with the v4 API.
type ClusterDefinitionV4 struct {
	Name              string            `yaml:"name,omitempty"`
	Owner             string            `yaml:"owner,omitempty"`
	ReleaseVersion    string            `yaml:"release_version,omitempty"`
	AvailabilityZones int               `yaml:"availability_zones,omitempty"`
	Scaling           ScalingDefinition `yaml:"scaling,omitempty"`
	Workers           []NodeDefinition  `yaml:"workers,omitempty"`
}

// ClusterDefinitionV5 defines a tenant cluster spec compatible with the v5 API.
type ClusterDefinitionV5 struct {
	APIVersion     string                `yaml:"api_version,omitempty"`
	Name           string                `yaml:"name,omitempty"`
	Owner          string                `yaml:"owner,omitempty"`
	ReleaseVersion string                `yaml:"release_version,omitempty"`
	Master         *MasterDefinition     `yaml:"master,omitempty"`
	MasterNodes    *MasterNodes          `yaml:"master_nodes,omitempty"`
	NodePools      []*NodePoolDefinition `yaml:"nodepools,omitempty"`
	Labels         map[string]*string    `yaml:"labels,omitempty"`
}

// ScalingDefinition defines how a tenant cluster can scale.
type ScalingDefinition struct {
	Min int64 `yaml:"min,omitempty"`
	Max int64 `yaml:"max,omitempty"`
}

// MasterDefinition defines a master in cluster creation, as introduced by the V5 API.
type MasterDefinition struct {
	AvailabilityZone string `yaml:"availability_zone,omitempty"`
}

// MasterNodes defines an interface for configuring HA master nodes.
type MasterNodes struct {
	HighAvailability bool `yaml:"high_availability,omitempty"`
}

// AvailabilityZonesDefinition defines the availability zones for a node pool, as intgroduc ed in the V5 API.
type AvailabilityZonesDefinition struct {
	Number int64    `yaml:"number,omitempty"`
	Zones  []string `yaml:"zones,omitempty"`
}

// NodeSpec defines the specification of the nodes in a node pool, as intriduced with the V5 API.
type NodeSpec struct {
	AWS *AWSSpecificDefinition `yaml:"aws,omitempty"`
}

// NodePoolDefinition defines a node pool as introduces by the V5 API.
type NodePoolDefinition struct {
	Name              string                       `yaml:"name,omitempty"`
	AvailabilityZones *AvailabilityZonesDefinition `yaml:"availability_zones,omitempty"`
	Scaling           *ScalingDefinition           `yaml:"scaling,omitempty"`
	NodeSpec          *NodeSpec                    `yaml:"node_spec,omitempty"`
}
