package flags

var (
	// CmdAPIEndpoint represents the API endpoint URL flag.
	CmdAPIEndpoint string

	// CmdToken represents the auth token passed as a flag.
	CmdToken string

	// CmdConfigDirPath represents the configuration path to use temporarily passed as a flag.
	CmdConfigDirPath string

	// CmdVerbose represents the verbosity switch passed as a flag.
	CmdVerbose bool

	// CmdCertificateOrganizations represents the O value for key pairs passed as a flag.
	CmdCertificateOrganizations string

	// CmdClusterID represents the cluster ID passed as a flag.
	CmdClusterID string

	// CmdCNPrefix represents the CN prefix passed as a flag.
	CmdCNPrefix string

	// CmdDescription represents the description passed as a flag.
	CmdDescription string

	// CmdTTL represents a TTL (time to live) value passed as a flag.
	CmdTTL string

	// CmdForce represents the value of the force flag, passed as a flag.
	// If true, all warnings should be suppressed.
	CmdForce bool

	// CmdFull represents the switch to disable all output truncation, passed as a flag.
	CmdFull bool

	// CmdNumWorkers is the number of workers required via flag on execution.
	CmdNumWorkers int

	// CmdOrganizationID represents an organization ID, passed as a flag.
	CmdOrganizationID string

	// CmdRelease sets a release to use, provided as a command line flag.
	CmdRelease string

	// CmdWorkerNumCPUs prepresents the number of CPUs per worker as required via flag.
	CmdWorkerNumCPUs int

	// CmdWorkerMemorySizeGB represents the RAM size per worker node in GB per worker as required via flag.
	CmdWorkerMemorySizeGB float32

	// CmdWorkerStorageSizeGB represents the local storage per worker node in GB per worker as required via flag.
	CmdWorkerStorageSizeGB float32

	// CmdWorkersMin is the minimum number of workers created for the cluster.
	CmdWorkersMin int64

	// CmdWorkersMax is the minimum number of workers created for the cluster.
	CmdWorkersMax int64
)
