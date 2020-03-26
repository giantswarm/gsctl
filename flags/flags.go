package flags

var (
	// APIEndpoint represents the API endpoint URL flag.
	APIEndpoint string

	// AvailabilityZones is the number of availability zones to use.
	AvailabilityZones int

	// ConfigDirPath represents the configuration path to use temporarily passed as a flag.
	ConfigDirPath string

	// Verbose represents the verbosity switch passed as a flag.
	Verbose bool

	// CertificateOrganizations represents the O value for key pairs passed as a flag.
	CertificateOrganizations string

	// ClusterID represents the cluster ID passed as a flag.
	ClusterID string

	// ClusterName is the cluster name set via flag on execution
	ClusterName string

	// CNPrefix represents the CN prefix passed as a flag.
	CNPrefix string

	// CreateDefaultNodePool defines whether a default node pool should be created
	// in the case that none was defined in the cluster definition.
	CreateDefaultNodePool bool

	// Description represents the description passed as a flag.
	Description string

	// Use spot instances for a node pool
	EnableSpotInstances bool

	// Force represents the value of the force flag, passed as a flag.
	// If true, all warnings should be suppressed.
	Force bool

	// Full represents the switch to disable all output truncation, passed as a flag.
	Full bool

	// InputYAMLFile is the path to the input file used optionally as cluster definition
	InputYAMLFile string

	// Name is the name of a cluster or node pool.
	Name string

	// NumWorkers is the number of workers required via flag on execution.
	NumWorkers int

	// OrganizationID represents an organization ID, passed as a flag.
	OrganizationID string

	// Owner is the owner organization of the cluster as set via flag on execution.
	Owner string

	// Release sets a release to use, provided as a command line flag.
	Release string

	// TenantInternal represents the type of Kubernetes API endpoints
	// used to generate kubeconfig
	TenantInternal bool

	// Token represents the auth token passed as a flag.
	Token string

	// TTL represents a TTL (time to live) value passed as a flag.
	TTL string

	// WorkerAwsEc2InstanceType is the instance type name for nodes in AWS.
	WorkerAwsEc2InstanceType string

	// WorkerAzureVMSize is the Azure VmSize to use, provided as a command line flag.
	WorkerAzureVMSize string

	// WorkerMemorySizeGB represents the RAM size per worker node in GB per worker as required via flag.
	WorkerMemorySizeGB float32

	// WorkerNumCPUs prepresents the number of CPUs per worker as required via flag.
	WorkerNumCPUs int

	// WorkerStorageSizeGB represents the local storage per worker node in GB per worker as required via flag.
	WorkerStorageSizeGB float32

	// WorkersMin is the minimum number of workers created for the cluster or node pool.
	WorkersMin int64

	// WorkersMax is the minimum number of workers created for the cluster or node pool.
	WorkersMax int64
)
