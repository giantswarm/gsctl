package limits

// TODO: These limitzs should ideally come from the API.
// See https://github.com/giantswarm/gsctl/issues/155

var (
	// MinimumNumWorkers is the minimum number of workers a cluster must have.
	MinimumNumWorkers int = 0

	// MinimumWorkerNumCPUs is the minimum number of CPUs a worjer node must have.
	MinimumWorkerNumCPUs int = 1

	// MinimumWorkerMemorySizeGB is the minimum amount of memory a worker node must have.
	MinimumWorkerMemorySizeGB float32 = 1

	// MinimumWorkerStorageSizeGB is the minimum storage size a worker node must have.
	MinimumWorkerStorageSizeGB float32 = 1
)
