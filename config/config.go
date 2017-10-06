package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/microerror"

	yaml "gopkg.in/yaml.v2"
)

const (
	// ConfigFileType is the type of config file we use
	ConfigFileType = "yaml"

	// ConfigFileName is the name of the configuration file, without ending
	ConfigFileName = "config"

	// ProgramName is the name of this program
	ProgramName = "gsctl"

	// VersionCheckURL is the URL telling us what the latest gsctl version is
	VersionCheckURL = "https://downloads.giantswarm.io/gsctl/VERSION"

	// VersionCheckInterval is the minimum time to wait between two version checks
	VersionCheckInterval = time.Hour * 24
)

var (
	// Config is an object holding all configuration fields
	Config = configStruct{}

	// Version is the version number, to be set on build by the go linker
	Version string

	// BuildDate is the build date, to be set on build by the go linker
	BuildDate string

	// Commit is the latest git commit hash, to be set on build by the go linker
	Commit string

	// HomeDirPath is the path to the user's home directory
	HomeDirPath string

	// DefaultConfigDirPath is the default config dir path to use
	DefaultConfigDirPath string

	// ConfigDirPath is the actual path of the config dir
	ConfigDirPath string

	// CertsDirPath is the path of the directory holding certificates
	CertsDirPath string

	// ConfigFilePath is the path of the configuration file
	ConfigFilePath string

	// KubeConfigPaths is the path(s) of kubeconfig files as slice of strings
	KubeConfigPaths []string

	// SystemUser is the current system user as user.User (os/user)
	SystemUser *user.User
)

// configStruct is the top-level data structure used to serialize and
// deserialize our configuration from/to a YAML file
type configStruct struct {

	// LastVersionCheck is the last time when we successfully checked for a gsctl update.
	// It has no "omitempty", to enforce the output. Marshaling failed otherwise.
	LastVersionCheck time.Time `yaml:"last_version_check"`

	// Updated is the time when the config has last been written.
	Updated string `yaml:"updated"`

	// Endpoints is a map of endpoints
	Endpoints map[string]*endpointConfigStruct `yaml:"endpoints"`

	// SelectedEndpoint is the URL of the selected endpoint
	SelectedEndpoint string `yaml:"selected_endpoint"`

	// Token is the token found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Token string `yaml:"-"`

	// Email is the user email found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Email string `yaml:"-"`
}

// endpointConfigStruct is used to serialize/deserialize endpoint configuration
// to/from a config file
type endpointConfigStruct struct {
	// Email is the email address of the authenticated user.
	Email string `yaml:"email"`

	// Token is the session token of the authenticated user.
	Token string `yaml:"token"`
}

// StoreEndpointAuth adds an endpoint to the configStruct.Endpoints field
// (if not yet there). This should only be done after successful authentication.
func (c *configStruct) StoreEndpointAuth(endpointURL string, email string, token string) error {
	ep := normalizeEndpoint(endpointURL)

	if email == "" || token == "" {
		return microerror.Mask(credentialsRequiredError)
	}

	c.Endpoints[ep] = &endpointConfigStruct{
		Email: email,
		Token: token,
	}

	WriteToFile()

	return nil
}

// SelectEndpoint makes the given endpoint URL the selected one
func (c *configStruct) SelectEndpoint(endpointURL string) error {
	ep := normalizeEndpoint(endpointURL)
	if _, ok := c.Endpoints[ep]; !ok {
		return microerror.Mask(endpointNotDefinedError)
	}

	c.SelectedEndpoint = ep
	c.Token = c.Endpoints[ep].Token
	c.Email = c.Endpoints[ep].Email

	WriteToFile()

	return nil
}

// SelectedEndpoint returns the selected endpoint URL.
// If the argument overridingEndpointURL is not empty, this will
// be used as the returned endpoint URL.
// Errors are only printed to inform users, but not returned, to simplify
// usage of this function.
func (c *configStruct) ChooseEndpoint(overridingEndpointURL string) string {
	if overridingEndpointURL != "" {
		ep := normalizeEndpoint(overridingEndpointURL)
		return ep
	}

	envEndpoint := os.Getenv("GSCTL_ENDPOINT")
	if envEndpoint != "" {
		ep := normalizeEndpoint(envEndpoint)
		return ep
	}

	return c.SelectedEndpoint
}

// Logout removes the token value from the selected endpoint.
func (c *configStruct) Logout(endpointURL string) {
	ep := normalizeEndpoint(endpointURL)

	if ep == c.SelectedEndpoint {
		c.Token = ""
	}

	if element, ok := c.Endpoints[ep]; ok {
		element.Token = ""
	}

	WriteToFile()
}

// init sets defaults and initializes config paths
func init() {
	SystemUser, err := user.Current()
	if err != nil {
		fmt.Println("Could not get system user details using os/user.Current().")
		fmt.Printf("Without this information, %s cannot determine the user's home directory and cannot set a configuration path.\n", ProgramName)
		fmt.Println("Please get in touch with us via support@giantswarm.io, including information on your platform.")
		fmt.Println("Thank you and sorry for the inconvenience!")
		fmt.Println("")
		panic(err.Error())
	}
	HomeDirPath = SystemUser.HomeDir

	// create default config dir path
	DefaultConfigDirPath = path.Join(HomeDirPath, ".config", ProgramName)
}

// Initialize sets up all configuration.
// It's distinct from init() on purpose, so it's
// execution can be triggered in a controlled way.
// It's supposed to be called after init().
// The configDirPath argument can be given to override the DefaultConfigDirPath.
func Initialize(configDirPath string) error {
	// Reset our Config object. This is particularly necessary for running
	// multiple tests in a row.
	Config = configStruct{}

	// configDirPath argument overrides default, if given
	if configDirPath != "" {
		ConfigDirPath = configDirPath
	} else {
		ConfigDirPath = DefaultConfigDirPath
	}

	ConfigFilePath = path.Join(ConfigDirPath, ConfigFileName+"."+ConfigFileType)

	// if config file doesn't exist, create empty one
	_, err := os.Stat(ConfigFilePath)
	if os.IsNotExist(err) {
		// ensure directory exists
		dirErr := os.MkdirAll(ConfigDirPath, 0700)
		if dirErr != nil {
			return dirErr
		}
		// ensure file exists
		file, fileErr := os.Create(ConfigFilePath)
		if fileErr != nil {
			return fileErr
		}
		file.Close()
	}

	myConfig, err := readFromFile(ConfigFilePath)
	if err != nil {
		return microerror.Mask(err)
	}
	populateConfigStruct(myConfig)

	CertsDirPath = path.Join(ConfigDirPath, "certs")
	os.MkdirAll(CertsDirPath, 0700)

	KubeConfigPaths = getKubeconfigPaths(HomeDirPath)

	return nil
}

// populateConfigStruct assigns configuration values from the unmarshalled
// structure to Config.
// cs here is what we read from the file.
func populateConfigStruct(cs configStruct) {

	Config.LastVersionCheck = cs.LastVersionCheck
	Config.Updated = cs.Updated

	Config.Endpoints = cs.Endpoints
	if Config.Endpoints == nil {
		Config.Endpoints = make(map[string]*endpointConfigStruct)
	}

	if cs.SelectedEndpoint != "" {
		Config.SelectedEndpoint = cs.SelectedEndpoint
		if _, ok := cs.Endpoints[cs.SelectedEndpoint]; ok {
			Config.Email = cs.Endpoints[cs.SelectedEndpoint].Email
			Config.Token = cs.Endpoints[cs.SelectedEndpoint].Token
		}
	}

}

// UserAgent returns the user agent string identifying us in HTTP requests
func UserAgent() string {
	return fmt.Sprintf("%s/%s", ProgramName, Version)
}

// readFromFile reads configuration from the YAML config file
func readFromFile(filePath string) (configStruct, error) {
	myConfig := configStruct{}
	data, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// ignore if file does not exist,
			// as this is not an error.
			return myConfig, nil
		}
		return myConfig, microerror.Mask(readErr)
	}

	yamlErr := yaml.Unmarshal(data, &myConfig)
	if yamlErr != nil {
		return myConfig, microerror.Mask(yamlErr)
	}

	return myConfig, nil
}

// WriteToFile writes the configuration data to a YAML file
func WriteToFile() error {

	data := Config
	data.Updated = time.Now().Format(time.RFC3339)

	yamlBytes, err := yaml.Marshal(&data)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ioutil.WriteFile(ConfigFilePath, yamlBytes, 0600)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// getKubeconfigPaths returns a slice of paths to known kubeconfig files
func getKubeconfigPaths(homeDir string) []string {
	// check if KUBECONFIG environment variable is set
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv != "" {
		// KUBECONFIG is set.
		// Now check all the paths included for file existence

		paths := strings.Split(kubeconfigEnv, string(os.PathListSeparator))
		out := []string{}
		for _, myPath := range paths {
			if _, err := os.Stat(myPath); err == nil {
				out = append(out, myPath)
			}
		}
		return out
	}

	// KUBECONFIG is not set.
	// Look for the default location ~/.kube/config
	filePath := path.Join(homeDir, ".kube", "config")

	if _, err := os.Stat(filePath); err == nil {
		// file exists in default location
		return []string{filePath}
	}

	// No kubeconfig file. Return empty slice.
	return nil
}

// GetDefaultCluster determines which is the default cluster
//
// This can be either the only cluster accessible, or a cluster selected explicitly.
//
// @param requestIDHeader  Request ID to pass with API requests
// @param activityName     Name of the activity calling this function (for tracking)
// @param cmdLine          Command line content used to run the CLI (for tracking)
// @param apiEndpoint      Endpoint URL
func GetDefaultCluster(requestIDHeader, activityName, cmdLine, apiEndpoint string) (clusterID string, err error) {
	// Go through available orgs and clusters to find all clusters
	if Config.Token == "" {
		return "", errors.New("user not logged in")
	}

	clientConfig := client.Configuration{
		Endpoint:  apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		return "", microerror.Mask(clientErr)
	}

	authHeader := "giantswarm " + Config.Token

	clustersResponse, _, err := apiClient.GetClusters(authHeader, requestIDHeader, activityName, cmdLine)
	if err != nil {
		return "", err
	}

	if len(clustersResponse) == 1 {
		return clustersResponse[0].Id, nil
	}

	return "", nil
}

// normalizeEndpoint sanitizes a user-entered endpoint URL.
// - turn to lowercase
// - Adds https:// if no scheme is given
// - Removes path and other junk
// Errors are printed immediately here, to simplify handling of this function.
func normalizeEndpoint(u string) string {
	// lowercase
	u = strings.ToLower(u)

	// if URL has no scheme, prefix it with the default scheme
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		u = "https://" + u
	}

	// strip extra stuff
	p, err := url.Parse(u)
	if err != nil {
		fmt.Printf("[Warning] Endpoint URL normalization yielded: %s\n", err)
	}

	// remove everything but scheme and host
	u = p.Scheme + "://" + p.Host

	return u
}
