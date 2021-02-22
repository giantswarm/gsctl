package flags

var (
	// APIEndpoint represents the API endpoint URL flag.
	APIEndpoint string

	// AvailabilityZones is the number of availability zones to use.
	AvailabilityZones int

	// AWSUseAlikeInstanceTypes determines if similar instance types are used in an node pool
	AWSUseAlikeInstanceTypes bool

	// AWSOnDemandBaseCapacity determines the number of on-demand instances used in a node pool
	// before starting to use spot instances
	AWSOnDemandBaseCapacity int64

	// AWSSpotPercentage represents the percentage of spot instances
	// to use in an node pool
	AWSSpotPercentage int64

	// AzureSpotInstances determines if spot instances are used in a node pool.
	AzureSpotInstances bool

	// AzureSpotInstancesMaxPrice represents the maximum value that
	// a single node pool VM instance can reach before it is deallocated.
	AzureSpotInstancesMaxPrice float64

	// ConfigDirPath represents the configuration path to use temporarily passed as a flag.
	ConfigDirPath string

	// InternalAPI is a flag that causes the 'create kubeconfig' and 'create keypair'
	// command to use the workload-cluster-internal API endpoint instead of the public one.
	InternalAPI bool

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

	// UseKubie is used to set the context with Kubie
	UseKubie bool

	// Label contains label changes passed as multiple flags.
	Label []string

	// Name is the name of a cluster or node pool.
	Name string

	// NumWorkers is the number of workers required via flag on execution.
	NumWorkers int

	// OrganizationID represents an organization ID, passed as a flag.
	OrganizationID string

	// OutputFormat is the output format (table or json) of a commands output, passed as a flag.
	OutputFormat string

	// Owner is the owner organization of the cluster as set via flag on execution.
	Owner string

	// Release sets a release to use, provided as a command line flag.
	Release string

	// SilenceHTTPEndpointWarning represents
	SilenceHTTPEndpointWarning bool

	// MasterHA enables or disabled master node high availability.
	MasterHA bool

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

	// WorkersMin is used to detect if the parameter has been set in the commandline
	WorkersMinSet bool

	// WorkersMax is the minimum number of workers created for the cluster or node pool.
	WorkersMax int64
)
