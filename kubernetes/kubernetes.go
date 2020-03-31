package kubernetes

type KubeConfigValue struct {
	APIVersion     string                   `yaml:"apiVersion,omitempty"`
	Kind           string                   `yaml:"kind,omitempty"`
	Clusters       []KubeconfigNamedCluster `yaml:"clusters,omitempty"`
	Users          []KubeconfigUser         `yaml:"users,omitempty"`
	Contexts       []KubeconfigNamedContext `yaml:"contexts,omitempty"`
	CurrentContext string                   `yaml:"current-context,omitempty"`
	Preferences    struct{}                 `yaml:"preferences,omitempty"`
}

// KubeconfigUser is a struct used to create a kubectl configuration YAML file
type KubeconfigUser struct {
	Name string                `yaml:"name,omitempty"`
	User KubeconfigUserKeyPair `yaml:"user,omitempty"`
}

// KubeconfigUserKeyPair is a struct used to create a kubectl configuration YAML file
type KubeconfigUserKeyPair struct {
	ClientCertificateData string                 `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string                 `yaml:"client-key-data,omitempty"`
	AuthProvider          KubeconfigAuthProvider `yaml:"auth-provider,omitempty,omitempty"`
}

// KubeconfigAuthProvider is a struct used to create a kubectl authentication provider
type KubeconfigAuthProvider struct {
	Name   string            `yaml:"name,omitempty"`
	Config map[string]string `yaml:"config,omitempty"`
}

// KubeconfigNamedCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedCluster struct {
	Name    string            `yaml:"name,omitempty"`
	Cluster KubeconfigCluster `yaml:"cluster,omitempty"`
}

// KubeconfigCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigCluster struct {
	Server                   string `yaml:"server,omitempty"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
	CertificateAuthority     string `yaml:"certificate-authority,omitempty"`
}

// KubeconfigNamedContext is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedContext struct {
	Name    string            `yaml:"name,omitempty"`
	Context KubeconfigContext `yaml:"context,omitempty"`
}

// KubeconfigContext is a struct used to create a kubectl configuration YAML file
type KubeconfigContext struct {
	Cluster   string `yaml:"cluster,omitempty"`
	Namespace string `yaml:"namespace,omitempty,omitempty"`
	User      string `yaml:"user,omitempty"`
}
