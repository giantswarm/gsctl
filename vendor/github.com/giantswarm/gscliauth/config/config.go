package config

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/giantswarm/gscliauth/oidc"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const (
	// ConfigFileType is the type of config file we use.
	ConfigFileType = "yaml"

	// ConfigFileName is the name of the configuration file, without ending.
	ConfigFileName = "config"

	// ConfigFilePermission is the rights mask for the config file.
	ConfigFilePermission = 0600

	// ProgramName is the name of this program.
	ProgramName = "gsctl"

	// VersionCheckURL is the URL telling us what the latest gsctl version is.
	VersionCheckURL = "https://github.com/giantswarm/gsctl/releases/latest"

	// VersionCheckInterval is the minimum time to wait between two version checks.
	VersionCheckInterval = time.Hour * 24

	// garbageCollectionLikelihood is a number between 0 and 1 that sets the
	// likelihood that we will execute garbage collection functions..
	garbageCollectionLikelihood = .2
)

var (
	// Config is an object holding all configuration fields.
	Config = newConfigStruct()

	// Version is the version number, to be set on build by the go linker.
	Version string

	// BuildDate is the build date, to be set on build by the go linker.
	BuildDate string

	// Commit is the latest git commit hash, to be set on build by the go linker.
	Commit string

	// FileSystem is the afero filesystem used to store the config file and certificates.
	FileSystem afero.Fs

	// HomeDirPath is the path to the user's home directory.
	HomeDirPath string

	// DefaultConfigDirPath is the default config dir path to use.
	DefaultConfigDirPath string

	// ConfigDirPath is the actual path of the config dir.
	ConfigDirPath string

	// CertsDirPath is the path of the directory holding certificates.
	CertsDirPath string

	// ConfigFilePath is the path of the configuration file.
	ConfigFilePath string

	// KubeConfigPaths is the path(s) of kubeconfig files as slice of strings.
	KubeConfigPaths []string

	// SystemUser is the current system user as user.User (os/user).
	SystemUser *user.User
)

// configStruct is the top-level data structure used to serialize and
// deserialize our configuration from/to a YAML file.
type configStruct struct {
	// LastVersionCheck is the last time when we successfully checked for a gsctl update.
	// It has no "omitempty", to enforce the output. Marshaling failed otherwise.
	LastVersionCheck time.Time `yaml:"last_version_check"`

	// Updated is the time when the config has last been written.
	Updated string `yaml:"updated"`

	// SelectedEndpoint is the URL of the selected endpoint.
	SelectedEndpoint string `yaml:"selected_endpoint"`

	// RefreshToken is the refresh token found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	RefreshToken string `yaml:"-"`

	// Scheme is the scheme found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Scheme string `yaml:"-"`

	// Token is the token found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Token string `yaml:"-"`

	// Email is the user email found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Email string `yaml:"-"`

	// Provider is the provider found for the selected endpoint. Might be empty.
	// Not marshalled back to the config file, as it is contained in the
	// endpoint's entry.
	Provider string `yaml:"-"`

	endpoints      map[string]*endpointConfig
	endpointsMutex *sync.RWMutex
}

func newConfigStruct() *configStruct {
	return &configStruct{
		endpoints:      map[string]*endpointConfig{},
		endpointsMutex: &sync.RWMutex{},
	}
}

// readFromFile reads configuration from the YAML config file.
func readFromFile(fs afero.Fs, filePath string) (*configStruct, error) {
	config := newConfigStruct()

	doesExist, err := afero.Exists(fs, filePath)
	if err != nil {
		return config, microerror.Mask(err)
	}
	if !doesExist {
		return config, nil
	}

	data, err := afero.ReadFile(fs, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// ignore if file does not exist,
			// as this is not an error.
			return config, nil
		}
		return config, microerror.Mask(err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, microerror.Mask(err)
	}

	return config, nil
}

// endpointConfig is used to serialize/deserialize endpoint configuration
// to/from a config file.
type endpointConfig struct {
	// Alias is a friendly shortcut for the endpoint.
	Alias string `yaml:"alias,omitempty"`

	// Email is the email address of the authenticated user.
	Email string `yaml:"email"`

	// Provider is the cloud provider used in the installation.
	Provider string `yaml:"provider"`

	// RefreshToken for acquiring a new token when using the bearer scheme.
	RefreshToken string `yaml:"refresh_token,omitempty"`

	// Scheme is the scheme to be used in the Authorization header.
	Scheme string `yaml:"auth_scheme,omitempty"`

	// Token is the session token of the authenticated user.
	Token string `yaml:"token,omitempty"`
}

// StoreEndpointAuth adds an endpoint to the configStruct.Endpoints field
// (if not yet there). This should only be done after successful authentication.
func (c *configStruct) StoreEndpointAuth(endpointURL, alias, provider, email, scheme, token, refreshToken string) error {
	ep := normalizeEndpoint(endpointURL)

	if email == "" || token == "" {
		return microerror.Mask(credentialsRequiredError)
	}

	// Ensure alias uniqueness.
	// If the alias is already in use, it has to point to the
	// same endpoint URL.
	if alias != "" && c.HasEndpointAlias(alias) {
		aliasedURL, err := c.EndpointByAlias(alias)
		if err != nil {
			return microerror.Mask(err)
		}

		if aliasedURL != ep {
			return microerror.Mask(aliasMustBeUniqueError)
		}
	}

	c.endpointsMutex.Lock()
	defer c.endpointsMutex.Unlock()

	// keep current Alias, if there.
	aliasBefore := ""
	if _, ok := c.endpoints[ep]; ok {
		aliasBefore = c.endpoints[ep].Alias
	}

	if provider == "" && c.endpoints[ep] != nil {
		provider = c.endpoints[ep].Provider
	}

	c.endpoints[ep] = &endpointConfig{
		Alias:        aliasBefore,
		Email:        email,
		Provider:     provider,
		RefreshToken: refreshToken,
		Scheme:       scheme,
		Token:        token,
	}

	if alias != "" && aliasBefore == "" {
		c.endpoints[ep].Alias = alias
	}

	WriteToFile()

	return nil
}

// SelectEndpoint makes the given endpoint the selected one. The argument
// can either be an alias (that will be used as is) or
// a URL which will undergo normalization.
func (c *configStruct) SelectEndpoint(endpointAliasOrURL string) error {

	if endpointAliasOrURL == "" {
		return microerror.Mask(endpointNotDefinedError)
	}

	ep := ""

	argumentIsAlias := false

	// first check if the endpointURL matches an alias.
	if c.HasEndpointAlias(endpointAliasOrURL) {
		argumentIsAlias = true
		var epErr error
		ep, epErr = c.EndpointByAlias(endpointAliasOrURL)
		if epErr != nil {
			return microerror.Mask(epErr)
		}
	}

	c.endpointsMutex.Lock()
	c.endpointsMutex.Unlock()

	if !argumentIsAlias {
		ep = normalizeEndpoint(endpointAliasOrURL)
		if _, ok := c.endpoints[ep]; !ok {
			return microerror.Mask(endpointNotDefinedError)
		}
	}

	// Migrate empty scheme to 'giantswarm'.
	if c.endpoints[ep].Scheme == "" {
		c.endpoints[ep].Scheme = "giantswarm"
	}

	c.SelectedEndpoint = ep
	c.RefreshToken = c.endpoints[ep].RefreshToken
	c.Scheme = c.endpoints[ep].Scheme
	c.Token = c.endpoints[ep].Token
	c.Email = c.endpoints[ep].Email

	WriteToFile()

	return nil
}

// ChooseEndpoint makes a choice which should be the endpoint to use.
// If the argument overridingEndpointAliasOrURL is not empty, this will
// be used to look up an alias to return an endpoint for. If there is none,
// it will be the used endpoint URL.
func (c *configStruct) ChooseEndpoint(overridingEndpointAliasOrURL string) string {

	// if no local param is given, try the environment variable.
	if overridingEndpointAliasOrURL == "" {
		overridingEndpointAliasOrURL = os.Getenv("GSCTL_ENDPOINT")
	}

	if overridingEndpointAliasOrURL != "" {
		// check if overridingEndpointAliasOrURL is an alias.
		if c.HasEndpointAlias(overridingEndpointAliasOrURL) {
			ep, _ := c.EndpointByAlias(overridingEndpointAliasOrURL)
			return ep
		}

		ep := normalizeEndpoint(overridingEndpointAliasOrURL)
		return ep
	}

	// as a last resort, return the currently selected endpoint.
	return c.SelectedEndpoint
}

// ChooseToken chooses a token to use, according to a rule set.
// - If the given token is not empty, we use (return) that
// - If the given token is empty and we have an auth token for the given
//   endpoint, we return that
// - otherwise we return an empty string
func (c *configStruct) ChooseToken(endpoint, overridingToken string) string {
	ep := normalizeEndpoint(endpoint)

	if overridingToken != "" {
		return overridingToken
	}

	endpointConfig := c.EndpointConfig(ep)
	if endpointConfig != nil && endpointConfig.Token != "" {
		return endpointConfig.Token
	}

	return ""
}

// ChooseScheme chooses a scheme to use, according to a rule set.
// - If the user is providing their own token via the --auth-token flag,
//   then always return "giantswarm".
// - If we have an auth scheme for the given endpoint, we return that.
// - otherwise we return "giantswarm"
func (c *configStruct) ChooseScheme(endpoint string, CmdToken string) string {
	ep := normalizeEndpoint(endpoint)

	if CmdToken != "" {
		return "giantswarm"
	}

	endpointConfig := c.EndpointConfig(ep)
	if endpointConfig != nil && endpointConfig.Scheme != "" {
		return endpointConfig.Scheme
	}

	return "giantswarm"
}

// HasEndpointAlias returns whether the given alias is used for an endpoint.
func (c *configStruct) HasEndpointAlias(alias string) bool {
	c.endpointsMutex.RLock()
	defer c.endpointsMutex.RUnlock()

	for key := range c.endpoints {
		if c.endpoints[key].Alias == alias {
			return true
		}
	}
	return false
}

func (c *configStruct) EndpointConfig(ep string) *endpointConfig {
	c.endpointsMutex.RLock()
	defer c.endpointsMutex.RUnlock()

	return c.endpoints[ep]
}

// EndpointByAlias performs a lookup by alias and returns the according endpoint URL
// (if the alias is assigned) or an error (if not found).
func (c *configStruct) EndpointByAlias(alias string) (string, error) {
	c.endpointsMutex.RLock()
	defer c.endpointsMutex.RUnlock()

	for url := range c.endpoints {
		if c.endpoints[url].Alias == alias {
			return url, nil
		}
	}
	return "", microerror.Maskf(endpointNotDefinedError, "no endpoint for this alias")
}

// Endpoints returns a slice of endpoint URLs.
func (c *configStruct) Endpoints() []string {
	c.endpointsMutex.RLock()
	defer c.endpointsMutex.RUnlock()

	var endpoints []string
	for k := range c.endpoints {
		endpoints = append(endpoints, k)
	}

	return endpoints
}

// SetProvider sets the provider information for the current endpoint.
// This fails if a provider is already set.
func (c *configStruct) SetProvider(provider string) error {
	if c.SelectedEndpoint == "" {
		return microerror.Mask(noEndpointSelectedError)
	}
	if c.Provider != "" {
		return microerror.Mask(endpointProviderIsImmuttableError)
	}

	c.endpointsMutex.Lock()
	defer c.endpointsMutex.Unlock()

	c.endpoints[c.SelectedEndpoint].Provider = provider
	c.Provider = provider
	WriteToFile()

	return nil
}

// NumEndpoints returns the number of endpoints stored in the configuration.
func (c *configStruct) NumEndpoints() int {
	c.endpointsMutex.RLock()
	defer c.endpointsMutex.RUnlock()

	return len(c.endpoints)
}

// Logout removes the token value from the selected endpoint.
func (c *configStruct) Logout(endpointURL string) {
	ep := normalizeEndpoint(endpointURL)

	if ep == c.SelectedEndpoint {
		c.Token = ""
		c.Scheme = ""
	}

	endpointConfig := c.EndpointConfig(ep)
	if endpointConfig != nil {
		endpointConfig.RefreshToken = ""
		endpointConfig.Token = ""
		endpointConfig.Scheme = ""
	}

	WriteToFile()
}

// AuthHeaderGetter returns a function that can get the auth header for a given endpoint that the client can use.
// The returned function will attempt to refresh the token in case the scheme is Bearer and the token is expired.
func (c *configStruct) AuthHeaderGetter(endpoint string, overridingToken string) func() (authheader string, err error) {
	return func() (string, error) {
		token := c.ChooseToken(endpoint, overridingToken)
		scheme := c.ChooseScheme(endpoint, overridingToken)

		// If the scheme is Bearer, first verify that the token is valid.
		// If it is expired, then try to refresh it.
		if scheme == "Bearer" {
			endpointConfig := c.EndpointConfig(endpoint)
			if endpointConfig == nil {
				return "", microerror.Mask(endpointNotDefinedError)
			}

			// Check if it has a refresh token.
			refreshToken := endpointConfig.RefreshToken
			if refreshToken == "" {
				return "", microerror.Maskf(endpointNotDefinedError, "No refresh token saved in config file, unable to acquire new access token. Please login again.")
			}

			if !isTokenValid(token) {
				// Get a new token.
				refreshTokenResponse, err := oidc.RefreshToken(refreshToken)
				if err != nil {
					return "", microerror.Mask(err)
				}

				// Parse the ID Token for the email address.
				idToken, err := oidc.ParseIDToken(refreshTokenResponse.IDToken)
				if err != nil {
					return "", microerror.Mask(err)
				}

				// Update the config file with the new access token.
				if err := Config.StoreEndpointAuth(endpoint, endpointConfig.Alias, "", idToken.Email, "Bearer", refreshTokenResponse.AccessToken, refreshToken); err != nil {
					return "", microerror.Maskf(err, "Error while attempting to store the token in the config file")
				}

				// Use the new access token.
				return scheme + " " + refreshTokenResponse.AccessToken, nil
			}
			return scheme + " " + token, nil
		}

		// If the scheme is not Bearer, just return scheme and token as normal.
		return scheme + " " + token, nil
	}
}

// isTokenValid takes a JWT access token and returns true/false depending on
// whether or not the access token is valid. Does not check if the signature is valid.
// Only checkes time based claims.
func isTokenValid(token string) (expired bool) {
	// Parse token
	claims := jwt.MapClaims{}

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, claims)
	if err != nil {
		return false
	}

	err = parsedToken.Claims.Valid()
	if err != nil {
		return false
	}

	return true
}

// init sets defaults and initializes config paths.
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
func Initialize(fs afero.Fs, configDirPath string) error {
	FileSystem = fs
	// Reset our Config object. This is particularly necessary for running
	// multiple tests in a row.
	Config = newConfigStruct()

	// configDirPath argument overrides default, if given.
	if configDirPath != "" {
		ConfigDirPath = configDirPath
	} else {
		ConfigDirPath = DefaultConfigDirPath
	}

	ConfigFilePath = path.Join(ConfigDirPath, ConfigFileName+"."+ConfigFileType)

	// if config file doesn't exist, create empty one.
	exists, err := afero.Exists(fs, ConfigFilePath)
	if err != nil {
		return microerror.Mask(err)
	}
	if !exists {
		// ensure directory exists.
		err := fs.MkdirAll(ConfigDirPath, 0700)
		if err != nil {
			return microerror.Mask(err)
		}
		// ensure file exists.
		file, err := fs.Create(ConfigFilePath)
		if err != nil {
			return microerror.Mask(err)
		}
		err = file.Close()
		if err != nil {
			return microerror.Mask(err)
		}

		err = fs.Chmod(ConfigFilePath, ConfigFilePermission)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	myConfig, err := readFromFile(fs, ConfigFilePath)
	if err != nil {
		return microerror.Mask(err)
	}
	populateConfigStruct(myConfig)

	CertsDirPath = path.Join(ConfigDirPath, "certs")
	err = fs.MkdirAll(CertsDirPath, 0700)
	if err != nil {
		return microerror.Mask(err)
	}

	KubeConfigPaths = getKubeconfigPaths(HomeDirPath)

	// apply garbage collection.
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	if randGenerator.Float32() < garbageCollectionLikelihood {
		err := GarbageCollectKeyPairs(fs)
		if err != nil {
			// print error message, but don't interrupt the user.
			if IsGarbageCollectionFailedError(err) {
				fmt.Printf("Error in key pair garbage collection - no files deleted: %s\n", err.Error())
			} else if IsGarbageCollectionPartiallyFailedError(err) {
				fmt.Printf("Error in key pair garbage collection - some files not deleted: %s\n", err.Error())
			} else {
				fmt.Printf("Error in key pair garbage collection: %s\n", err.Error())
			}
		}
	}

	return nil
}

// populateConfigStruct assigns configuration values from the unmarshalled
// structure to Config.
// cs here is what we read from the file.
func populateConfigStruct(cs *configStruct) {

	Config.LastVersionCheck = cs.LastVersionCheck
	Config.Updated = cs.Updated

	Config.endpoints = cs.endpoints

	if cs.SelectedEndpoint != "" {
		Config.SelectedEndpoint = cs.SelectedEndpoint
		endpointConfig := cs.EndpointConfig(cs.SelectedEndpoint)
		if endpointConfig != nil {
			Config.Email = endpointConfig.Email
			Config.Provider = endpointConfig.Provider
			Config.Token = endpointConfig.Token
		}
	}

}

// UserAgent returns the user agent string identifying us in HTTP requests.
func UserAgent() string {
	return fmt.Sprintf("%s/%s", ProgramName, Version)
}

// WriteToFile writes the configuration data to a YAML file.
func WriteToFile() error {
	data := Config
	data.Updated = time.Now().Format(time.RFC3339)

	yamlBytes, err := yaml.Marshal(&data)
	if err != nil {
		return microerror.Mask(err)
	}

	err = afero.WriteFile(FileSystem, ConfigFilePath, yamlBytes, ConfigFilePermission)
	if err != nil {
		return microerror.Mask(err)
	}

	// finally update permissions, in case they weren't right before.
	err = FileSystem.Chmod(ConfigFilePath, ConfigFilePermission)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// getKubeconfigPaths returns a slice of paths to known kubeconfig files.
func getKubeconfigPaths(homeDir string) []string {
	// check if KUBECONFIG environment variable is set
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv != "" {
		// Now check all the paths included for file existence.
		paths := strings.Split(kubeconfigEnv, string(os.PathListSeparator))
		var out []string
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

// normalizeEndpoint sanitizes a user-entered endpoint URL.
// - turn to lowercase
// - Adds https:// if no scheme is given
// - Removes path and other junk
// Errors are printed immediately here, to simplify handling of this function.
func normalizeEndpoint(u string) string {
	// lowercase.
	u = strings.ToLower(u)

	// if URL has no scheme, prefix it with the default scheme.
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		u = "https://" + u
	}

	// strip extra stuff.
	p, err := url.Parse(u)
	if err != nil {
		fmt.Printf("[Warning] Endpoint URL normalization yielded: %s\n", err)
	}

	// remove everything but scheme and host.
	u = p.Scheme + "://" + p.Host

	return u
}
