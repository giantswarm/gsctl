package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/giantswarm/gsctl/client"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

const (
	// ConfigFileType is the type of config file we use
	ConfigFileType = "yaml"

	// ConfigFileName is the name of the configuration file, without ending
	ConfigFileName = "config"

	// ProgramName is the name of this program
	ProgramName = "gsctl"

	// DefaultAPIEndpoint is the endpoint used if none is configured
	DefaultAPIEndpoint = "https://api.giantswarm.io"
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

	// LegacyConfigDirPath is the path where we had the config stuff earlier
	LegacyConfigDirPath string

	// CertsDirPath is the path of the directory holding certificates
	CertsDirPath string

	// ConfigFilePath is the path of the configuration file
	ConfigFilePath string

	// KubeConfigPaths is the path(s) of kubeconfig files as slice of strings
	KubeConfigPaths []string

	// SystemUser is the current system user as user.User (os/user)
	SystemUser *user.User
)

// configStruct is used to serialize our configuration back into a file
type configStruct struct {
	Token   string `yaml:"token,omitempty"`
	Email   string `yaml:"email,omitempty"`
	Updated string `yaml:"updated,omitempty"`
}

// init sets up viper and sets defaults
func init() {
	viper.SetConfigType(ConfigFileType)
	viper.SetConfigName(ConfigFileName)
	viper.SetEnvPrefix(ProgramName)
	viper.SetTypeByDefaultValue(true)

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
	LegacyConfigDirPath = path.Join(HomeDirPath, "."+ProgramName)
}

// Initialize sets up all configuration.
// It's distinct from init() on purpose, so it's
// execution can be triggered in a controlled way.
// It's supposed to be called after init().
// The configDirPath argument can be given to override the DefaultConfigDirPath.
func Initialize(configDirPath string) error {

	// configDirPath argument overrides default, if given
	if configDirPath != "" {
		ConfigDirPath = configDirPath
	} else {
		ConfigDirPath = DefaultConfigDirPath
	}
	viper.AddConfigPath(ConfigDirPath)

	ConfigFilePath = path.Join(ConfigDirPath, ConfigFileName+"."+ConfigFileType)

	// 2017-05-11: move legacy config from old to new default config path,
	// if present. This is independent of the config path actually applied by the
	// user.
	// TODO: remove this after a couple of months.
	migrateConfigDir()

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

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	populateConfigStruct()

	CertsDirPath = path.Join(ConfigDirPath, "certs")
	os.MkdirAll(CertsDirPath, 0700)

	KubeConfigPaths = getKubeconfigPaths(HomeDirPath)

	return nil
}

// populateConfigStruct assigns configuration values from viper to Config
func populateConfigStruct() {
	if viper.IsSet("email") {
		Config.Email = viper.GetString("email")
	}
	if viper.IsSet("token") {
		Config.Token = viper.GetString("token")
	}
}

// UserAgent returns the user agent string identifying us in HTTP requests
func UserAgent() string {
	return fmt.Sprintf("%s/%s", ProgramName, Version)
}

// WriteToFile writes the configuration data to a YAML file
func WriteToFile() error {

	data := Config
	data.Updated = time.Now().Format(time.RFC3339)

	yamlBytes, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(ConfigFilePath, yamlBytes, 0600)
	if err != nil {
		return err
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
func GetDefaultCluster(requestIDHeader, activityName, cmdLine, cmdAPIEndpoint string) (clusterID string, err error) {
	// Go through available orgs and clusters to find all clusters
	if Config.Token == "" {
		return "", errors.New("user not logged in")
	}

	clientConfig := client.Configuration{
		Endpoint:  cmdAPIEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)

	authHeader := "giantswarm " + Config.Token
	orgsResponse, _, err := apiClient.GetUserOrganizations(authHeader, requestIDHeader, activityName, cmdLine)
	if err != nil {
		return "", err
	}
	if orgsResponse.StatusCode == 10000 {
		if len(orgsResponse.Data) > 0 {
			clusterIDs := []string{}
			for _, orgName := range orgsResponse.Data {
				clustersResponse, _, err := apiClient.GetOrganizationClusters(authHeader, orgName, requestIDHeader, activityName, cmdLine)
				if err != nil {
					return "", err
				}
				for _, cluster := range clustersResponse.Data.Clusters {
					clusterIDs = append(clusterIDs, cluster.Id)
				}
			}
			if len(clusterIDs) == 1 {
				return clusterIDs[0], nil
			}
			return "", nil
		}
	}
	return "", errors.New(orgsResponse.StatusText)
}

// migrateConfigDir migrates a configuration directory from the old
// default path to the new default path. Conditions:
// - old config dir exists
// - new config dir does not exist
func migrateConfigDir() error {
	_, err := os.Stat(DefaultConfigDirPath)
	if !os.IsNotExist(err) {
		// new config dir already exists
		return nil
	}

	_, err = os.Stat(LegacyConfigDirPath)
	if os.IsNotExist(err) {
		// old config dir does not exist
		return nil
	}

	// ensure ~/.config exists
	os.MkdirAll(path.Dir(DefaultConfigDirPath), 0700)

	err = os.Rename(LegacyConfigDirPath, DefaultConfigDirPath)
	if err != nil {
		return err
	}

	// adapt certificate paths in kubeconfig
	if len(KubeConfigPaths) > 0 {
		// we only adapt the first file found (on purpose)
		if stat, err := os.Stat(KubeConfigPaths[0]); err == nil {
			oldConfig, configErr := ioutil.ReadFile(KubeConfigPaths[0])
			if configErr != nil {
				return configErr
			}
			newConfig := strings.Replace(string(oldConfig), LegacyConfigDirPath, DefaultConfigDirPath, -1)
			// write back
			writeErr := ioutil.WriteFile(KubeConfigPaths[0], []byte(newConfig), stat.Mode())
			if writeErr != nil {
				return fmt.Errorf("Could not overwrite kubectl config file. %s", writeErr.Error())
			}
		}
	}

	return nil
}
