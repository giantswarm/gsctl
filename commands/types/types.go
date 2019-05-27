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
	InstanceType string `yaml:"instance_type,omitempty"`
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

// ClusterDefinition defines a tenant cluster spec.
type ClusterDefinition struct {
	Name              string            `yaml:"name,omitempty"`
	Owner             string            `yaml:"owner,omitempty"`
	ReleaseVersion    string            `yaml:"release_version,omitempty"`
	AvailabilityZones int               `yaml:"availability_zones,omitempty"`
	Scaling           ScalingDefinition `yaml:"scaling,omitempty"`
	Workers           []NodeDefinition  `yaml:"workers,omitempty"`
}

// ScalingDefinition defines how a tenant cluster can scale.
type ScalingDefinition struct {
	Min int64 `yaml:"min,omitempty"`
	Max int64 `yaml:"max,omitempty"`
}
