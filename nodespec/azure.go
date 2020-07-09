package nodespec

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"
)

var (
	// specYAML is the raw data on all necessary Azure VM sizes taken from
	// https://github.com/giantswarm/installations/blob/master/default-draughtsman-configmap-values.yaml
	// Warning: YAML in Golang is super fragile. There must not be any tabs in this string, otherwise
	// the marshalling will fail. However we will likely detect this in CI when running tests.
	azureInstanceTypesYAML = `
Standard_A2_v2:
  description: Av2-series, general purpose, 100 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 4
  memoryInMb: 4294.967296
  name: Standard_A2_v2
  numberOfCores: 2
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 21474.83648
Standard_A4_v2:
  description: Av2-series, general purpose, 100 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 8
  memoryInMb: 8589.934592
  name: Standard_A4_v2
  numberOfCores: 4
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 42949.67296
Standard_A8_v2:
  description: Av2-series, general purpose, 100 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 17179.869184
  name: Standard_A8_v2
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 85899.34592
Standard_D2_v3:
  description: Dv3-series, general purpose, 160-190 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 4
  memoryInMb: 8589.934592
  name: Standard_D2_v3
  numberOfCores: 2
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 53687.0912
Standard_D4_v3:
  description: Dv3-series, general purpose, 160-190 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 8
  memoryInMb: 17179.869184
  name: Standard_D4_v3
  numberOfCores: 4
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 107374.1824
Standard_D8_v3:
  description: Dv3-series, general purpose, 160-190 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 34359.738368
  name: Standard_D8_v3
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 21474.83648
Standard_D16_v3:
  description: Dv3-series, general purpose, 160-190 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 68719.476736
  name: Standard_D16_v3
  numberOfCores: 16
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 42949.67296
Standard_D32_v3:
  description: Dv3-series, general purpose, 160-190 ACU, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 137438.953472
  name: Standard_D32_v3
  numberOfCores: 32
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 85899.34592
Standard_D2s_v3:
  description: Dsv3-series, general purpose, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 4
  memoryInMb: 8589.934592
  name: Standard_D2s_v3
  numberOfCores: 2
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 17179.869184
Standard_D4s_v3:
  description: Dsv3-series, general purpose, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 8
  memoryInMb: 17179.869184
  name: Standard_D4s_v3
  numberOfCores: 4
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 34359.738368
Standard_D8s_v3:
  description: Dsv3-series, general purpose, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 34359.738368
  name: Standard_D8s_v3
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 68719.476736
Standard_D16s_v3:
  description: Dsv3-series, general purpose, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 68719.476736
  name: Standard_D16s_v3
  numberOfCores: 16
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 137438.953472
Standard_D32s_v3:
  description: Dsv3-series, general purpose, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 137438.953472
  name: Standard_D32s_v3
  numberOfCores: 32
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 274877.906944
Standard_E4s_v3:
  description: Esv3-series, memory optimized, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 8
  memoryInMb: 34359.738368
  name: Standard_E4s_v3
  numberOfCores: 4
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 68719.476736
Standard_E8a_v4:
  description: The Eav4-series utilize the 2.35Ghz AMD EPYCTM 7452 processor, no premium storage
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 68719.476736
  name: Standard_E8a_v4
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 137438.953472
Standard_E8as_v4:
  description: Easv4-series sizes utilize the 2.35Ghz AMD EPYCTM 7452 processor, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 68719.476736
  name: Standard_E8as_v4
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 137438.953472
Standard_E8s_v3:
  description: Esv3-series, memory optimized, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 68719.476736
  name: Standard_E8s_v3
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 137438.953472
Standard_E16s_v3:
  descriptions: Esv3-series, memory optimized, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 137438.953472
  name: Standard_E16s_v3
  numberOfCores: 16
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 274877.906944
Standard_E32s_v3:
  description: Esv3-series, memory optimized, 160-190 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 274877.906944
  name: Standard_E32s_v3
  numberOfCores: 32
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 549755.813888
Standard_F4s_v2:
  description: Fsv2-series, compute optimized, 195-210 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 8
  memoryInMb: 8589.934592
  name: Standard_F4s_v2
  numberOfCores: 4
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 34359.738368
Standard_F8s_v2:
  description: Fsv2-series, compute optimized, 195-210 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 16
  memoryInMb: 17179.869184
  name: Standard_F8s_v2
  numberOfCores: 8
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 68719.476736
Standard_F16s_v2:
  description: Fsv2-series, compute optimized, 195-210 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 34359.738368
  name: Standard_F16s_v2
  numberOfCores: 16
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 137438.953472
Standard_F32s_v2:
  description: Fsv2-series, compute optimized, 195-210 ACU, premium storage supported
  additionalProperties: {}
  maxDataDiskCount: 32
  memoryInMb: 68719.476736
  name: Standard_F32s_v2
  numberOfCores: 32
  osDiskSizeInMb: 1047552
  resourceDiskSizeInMb: 274877.906944
`
)

// ProviderAzure contains all provider specific info
type ProviderAzure struct {
	vmSizes map[string]VMSize
}

type VMSize struct {
	Description          string  `yaml:"description"`
	MaxDataDiskCount     int     `yaml:"maxDataDiskCount"`
	MemoryInMB           float64 `yaml:"memoryInMb"`
	Name                 string  `yaml:"name"`
	NumberOfCores        int64   `yaml:"numberOfCores"`
	OSDiskSizeInMB       int64   `yaml:"osDiskSizeInMb"`
	ResourceDiskSizeInMB float64 `yaml:"resourceDiskSizeInMb"`
}

// NewAzureProvider initiates a new Azure provider with the information about VM sizes.
func NewAzureProvider() (*ProviderAzure, error) {
	p := &ProviderAzure{}

	err := yaml.Unmarshal([]byte(azureInstanceTypesYAML), &p.vmSizes)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return p, nil
}

// GetInstanceTypeDetails returns info on a certain instance type
func (p *ProviderAzure) GetVMSizeDetails(name string) (*VMSize, error) {
	vmSize, ok := p.vmSizes[name]
	if ok {
		return &vmSize, nil
	}

	fmt.Printf("VM size not found %q", name)

	return nil, microerror.Mask(vmSizeNotFoundErr)
}
