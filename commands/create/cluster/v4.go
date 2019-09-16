package cluster

import (
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/commands/types"
)

// readDefinitionFromYAMLV4 reads a cluster definition from YAML data
// that is compatible with the v4 cluster creation endpoint.
func readDefinitionFromYAMLV4(yamlBytes []byte) (*types.ClusterDefinitionV4, error) {
	def := &types.ClusterDefinitionV4{}

	err := yaml.Unmarshal(yamlBytes, def)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return def, nil
}

// readDefinitionFromFile reads a cluster definition from a YAML file
// that is compatible with the v4 cluster creation endpoint.
func readDefinitionFromFileV4(fs afero.Fs, path string) (*types.ClusterDefinitionV4, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return readDefinitionFromYAMLV4(data)
}

// updateDefinitionFromFlags extend/overwrites a clusterDefinition based on the
// flags/arguments the user has given.
func updateDefinitionFromFlagsV4(def *types.ClusterDefinitionV4, args Arguments) {
	if def == nil {
		return
	}

	if args.AvailabilityZones != 0 {
		def.AvailabilityZones = args.AvailabilityZones
	}

	if args.ClusterName != "" {
		def.Name = args.ClusterName
	}

	if args.ReleaseVersion != "" {
		def.ReleaseVersion = args.ReleaseVersion
	}

	if def.Scaling.Min > 0 && args.WorkersMin == 0 {
		args.WorkersMin = def.Scaling.Min
	}
	if def.Scaling.Max > 0 && args.WorkersMax == 0 {
		args.WorkersMax = def.Scaling.Max
	}

	if args.WorkersMax > 0 {
		def.Scaling.Max = args.WorkersMax
		args.NumWorkers = 1
		if args.WorkersMin == 0 {
			def.Scaling.Min = def.Scaling.Max
		}
	}
	if args.WorkersMin > 0 {
		def.Scaling.Min = args.WorkersMin
		args.NumWorkers = 1
		if args.WorkersMax == 0 {
			def.Scaling.Max = def.Scaling.Min
		}
	}

	if args.Owner != "" {
		def.Owner = args.Owner
	}

	if def.Scaling.Min == 0 && def.Scaling.Max == 0 {
		def.Scaling.Min = int64(args.NumWorkers)
		def.Scaling.Max = int64(args.NumWorkers)
	}

	if def.Scaling.Min == 0 && def.Scaling.Max == 0 && args.NumWorkers == 0 {
		def.Scaling.Min = 3
		def.Scaling.Max = 3
	}

	workers := []types.NodeDefinition{}

	worker := types.NodeDefinition{}
	if args.WorkerNumCPUs != 0 {
		worker.CPU = types.CPUDefinition{Cores: args.WorkerNumCPUs}
	}
	if args.WorkerStorageSizeGB != 0 {
		worker.Storage = types.StorageDefinition{SizeGB: args.WorkerStorageSizeGB}
	}
	if args.WorkerMemorySizeGB != 0 {
		worker.Memory = types.MemoryDefinition{SizeGB: args.WorkerMemorySizeGB}
	}
	// AWS-specific
	if args.WorkerAwsEc2InstanceType != "" {
		worker.AWS.InstanceType = args.WorkerAwsEc2InstanceType
	}
	// Azure
	if args.WorkerAzureVMSize != "" {
		worker.Azure.VMSize = args.WorkerAzureVMSize
	}
	workers = append(workers, worker)

	def.Workers = workers
}

// createAddClusterBodyV4 creates a models.V4AddClusterRequest from cluster definition.
func createAddClusterBodyV4(d *types.ClusterDefinitionV4) *models.V4AddClusterRequest {
	a := &models.V4AddClusterRequest{}
	a.AvailabilityZones = int64(d.AvailabilityZones)
	a.Name = d.Name
	a.Owner = &d.Owner
	a.ReleaseVersion = d.ReleaseVersion
	a.Scaling = &models.V4AddClusterRequestScaling{
		Min: d.Scaling.Min,
		Max: d.Scaling.Max,
	}

	if len(d.Workers) == 1 {
		ndmWorker := &models.V4AddClusterRequestWorkersItems{}
		ndmWorker.Memory = &models.V4AddClusterRequestWorkersItemsMemory{SizeGb: float64(d.Workers[0].Memory.SizeGB)}
		ndmWorker.CPU = &models.V4AddClusterRequestWorkersItemsCPU{Cores: int64(d.Workers[0].CPU.Cores)}
		ndmWorker.Storage = &models.V4AddClusterRequestWorkersItemsStorage{SizeGb: float64(d.Workers[0].Storage.SizeGB)}
		ndmWorker.Labels = d.Workers[0].Labels
		ndmWorker.Aws = &models.V4AddClusterRequestWorkersItemsAws{InstanceType: d.Workers[0].AWS.InstanceType}
		ndmWorker.Azure = &models.V4AddClusterRequestWorkersItemsAzure{VMSize: d.Workers[0].Azure.VMSize}
		a.Workers = append(a.Workers, ndmWorker)
	}

	return a
}
