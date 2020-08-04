// Package nodespec gives us access to provider-specific node specification.
package nodespec

import (
	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"
)

var (
	// specYAML is the raw data on all necessary AWS instance types taken from
	// https://github.com/giantswarm/installations/blob/master/default-draughtsman-configmap-values.yaml
	// Warning: YAML in Golang is super fragile. There must not be any tabs in this string, otherwise
	// the marshalling will fail. However we will likely detect this in CI when running tests.
	awsInstanceTypesYAML = `
c5.large:
  cpu_cores: 2
  description: C5 Compute Optimized Large
  memory_size_gb: 4
  storage_size_gb: 0
c5.xlarge:
  cpu_cores: 4
  description: C5 Compute Optimized Extra Large
  memory_size_gb: 8
  storage_size_gb: 0
c5.2xlarge:
  cpu_cores: 8
  description: C5 Compute Optimized Double Extra Large
  memory_size_gb: 16
  storage_size_gb: 0
c5.4xlarge:
  cpu_cores: 16
  description: C5 Compute Optimized Quadruple Extra Large
  memory_size_gb: 32
  storage_size_gb: 0
c5.9xlarge:
  cpu_cores: 36
  description: C5 Compute Optimized Nonuple Extra Large
  memory_size_gb: 72
  storage_size_gb: 0
i3.xlarge:
  cpu_cores: 4
  description: I3 Storage Optimized Extra Large
  memory_size_gb: 30.5
  storage_size_gb: 950
m3.2xlarge:
  cpu_cores: 8
  description: M3 General Purpose Double Extra Large
  memory_size_gb: 30
  storage_size_gb: 80
m3.large:
  cpu_cores: 2
  description: M3 General Purpose Large
  memory_size_gb: 7.5
  storage_size_gb: 32
m3.xlarge:
  cpu_cores: 4
  description: M3 General Purpose Extra Large
  memory_size_gb: 15
  storage_size_gb: 40
m4.2xlarge:
  cpu_cores: 8
  description: M4 General Purpose Double Extra Large
  memory_size_gb: 32
  storage_size_gb: 0
m4.4xlarge:
  cpu_cores: 16
  description: M4 General Purpose Four Extra Large
  memory_size_gb: 64
  storage_size_gb: 0
m4.large:
  cpu_cores: 2
  description: M4 General Purpose Large
  memory_size_gb: 8
  storage_size_gb: 0
m4.xlarge:
  cpu_cores: 4
  description: M4 General Purpose Extra Large
  memory_size_gb: 16
  storage_size_gb: 0
r3.2xlarge:
  cpu_cores: 8
  description: R3 High-Memory Double Extra Large
  memory_size_gb: 61
  storage_size_gb: 160
r3.4xlarge:
  cpu_cores: 16
  description: R3 High-Memory Quadruple Extra Large
  memory_size_gb: 122
  storage_size_gb: 320
r3.8xlarge:
  cpu_cores: 32
  description: R3 High-Memory Eight Extra Large
  memory_size_gb: 244
  storage_size_gb: 320
r3.large:
  cpu_cores: 2
  description: R3 High-Memory Large
  memory_size_gb: 15.25
  storage_size_gb: 32
r3.xlarge:
  cpu_cores: 4
  description: R3 High-Memory Extra Large
  memory_size_gb: 30.5
  storage_size_gb: 80
r5.xlarge:
  cpu_cores: 4
  description: R5 High-Memory Extra Large
  memory_size_gb: 32
  storage_size_gb: 0
r5.2xlarge:
  cpu_cores: 8
  description: R5 High-Memory Double Extra Large
  memory_size_gb: 64
  storage_size_gb: 0
r5.4xlarge:
  cpu_cores: 16
  description: R5 High-Memory Quadruple Extra Large
  memory_size_gb: 128
  storage_size_gb: 0
r5.8xlarge:
  cpu_cores: 32
  description: R5 High-Memory Eight Extra Large
  memory_size_gb: 256
  storage_size_gb: 0
r5.12xlarge:
  cpu_cores: 48
  description: R5 High-Memory Twelve Extra Large
  memory_size_gb: 284
  storage_size_gb: 0
t2.2xlarge:
  cpu_cores: 8
  description: T2 General Purpose Double Extra Large
  memory_size_gb: 32
  storage_size_gb: 0
t2.large:
  cpu_cores: 2
  description: T2 General Purpose Large
  memory_size_gb: 8
  storage_size_gb: 0
t2.xlarge:
  cpu_cores: 4
  description: T2 General Purpose Extra Large
  memory_size_gb: 16
  storage_size_gb: 0
m5.large:
  cpu_cores: 2
  description: M5 General Purpose Large
  memory_size_gb: 8
  storage_size_gb: 0
m5.xlarge:
  cpu_cores: 4
  description: M5 General Purpose Extra Large
  memory_size_gb: 16
  storage_size_gb: 0
m5.2xlarge:
  cpu_cores: 8
  description: M5 General Purpose Double Extra Large
  memory_size_gb: 32
  storage_size_gb: 0
m5.4xlarge:
  cpu_cores: 16
  description: M5 General Purpose Quadruple Extra Large
  memory_size_gb: 64
  storage_size_gb: 0
m5.8xlarge:
  cpu_cores: 32
  description: M5 General Purpose 8x Extra Large
  memory_size_gb: 128
  storage_size_gb: 0
m5.12xlarge:
  cpu_cores: 48
  description: M5 General Purpose 12x Extra Large
  memory_size_gb: 192
  storage_size_gb: 0
m5.16xlarge:
  cpu_cores: 64
  description: M5 General Purpose 16x Extra Large
  memory_size_gb: 256
  storage_size_gb: 0
m5.24xlarge:
  cpu_cores: 96
  description: M5 General Purpose 24x Extra Large
  memory_size_gb: 384
  storage_size_gb: 0
p2.xlarge:
  cpu_cores: 4
  description: P2 Extra Large providing GPUs
  memory_size_gb: 61
  storage_size_gb: 0
p3.2xlarge:
  cpu_cores: 8
  description: P3 Double Extra Large providing GPUs
  memory_size_gb: 61
  storage_size_gb: 0
p3.8xlarge:
  cpu_cores: 32
  description: P3 Eight Extra Large providing GPUs
  memory_size_gb: 244
  storage_size_gb: 0
`
)

// ProviderAWS contains all provider specific info
type ProviderAWS struct {
	instanceTypes map[string]InstanceType
}

// InstanceType describes an AWS instance type
type InstanceType struct {
	CPUCores      int    `yaml:"cpu_cores"`
	Description   string `yaml:"description"`
	MemorySizeGB  int    `yaml:"memory_size_gb"`
	StorageSizeGB int    `yaml:"storage_size_gb"`
}

// NewAWS initiates a
func NewAWS() (*ProviderAWS, error) {
	p := &ProviderAWS{}

	err := yaml.Unmarshal([]byte(awsInstanceTypesYAML), &p.instanceTypes)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return p, nil
}

// GetInstanceTypeDetails returns info on a certain instance type
func (p *ProviderAWS) GetInstanceTypeDetails(name string) (*InstanceType, error) {
	instanceType, ok := p.instanceTypes[name]
	if ok {
		return &instanceType, nil
	}

	return nil, microerror.Mask(instanceTypeNotFoundErr)
}
