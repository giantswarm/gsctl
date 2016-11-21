package config

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsclientgen"
)

// struct used for YAML serialization of settings
type configStruct struct {
	Token        string
	Email        string
	Updated      string
	Organization string
	Cluster      string
}

const (
	// ConfigFileName is the name of the configuration file
	ConfigFileName string = "config.yaml"

	// ProgramName is the name of this program
	ProgramName string = "gsctl"
)

var (
	// Version is the version number, to be set on build by the go linker
	Version string

	// BuildDate is the build date, to be set on build by the go linker
	BuildDate string

	// Config holds the configuration variables
	Config *configStruct

	// ConfigDirPath is the path of the directory holding our config file
	ConfigDirPath string

	// ConfigFilePath is the path of the configuration file
	ConfigFilePath string

	// KubeConfigPaths is the path(s) of kubeconfig files as slice of strings
	KubeConfigPaths []string

	// SystemUser is the current system user as user.User (os/user)
	SystemUser *user.User
)

func init() {

	SystemUser, userErr := user.Current()
	checkErr(userErr)

	ConfigDirPath = SystemUser.HomeDir + string(os.PathSeparator) + "." + ProgramName
	ConfigFilePath = ConfigDirPath + string(os.PathSeparator) + ConfigFileName
	KubeConfigPaths, _ = getKubeconfigPaths(SystemUser.HomeDir)

	myConfig, err := readFromFile(ConfigFilePath)
	checkErr(err)
	Config = myConfig
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// WriteToFile writes the configuration data to a YAML file
func WriteToFile() error {

	// ensure directory
	os.MkdirAll(ConfigDirPath, 0700)

	// last modified date
	Config.Updated = time.Now().Format(time.RFC3339)

	yamlBytes, yamlErr := yaml.Marshal(&Config)
	if yamlErr != nil {
		return yamlErr
	}

	writeErr := ioutil.WriteFile(ConfigFilePath, yamlBytes, 0600)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

// ReadFromFile reads configuration from the YAML config file
func readFromFile(filePath string) (*configStruct, error) {
	myConfig := new(configStruct)

	data, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// ignore if file does not exist,
			// as this is not an error.
			return myConfig, nil
		}
		return nil, readErr
	}

	yamlErr := yaml.Unmarshal(data, &myConfig)
	if yamlErr != nil {
		return nil, yamlErr
	}

	return myConfig, nil
}

// getKubeconfigPaths returns a slice of paths to known kubeconfig files.
func getKubeconfigPaths(homeDir string) ([]string, error) {

	// check if KUBECONFIG environment variable is set
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv != "" {
		// KUBECONFIG is set.
		// Now check all the paths included for file existence

		// "windows" is really the value for the windows platform.
		// No constant for that in runtime.
		pathSep := ":"
		if runtime.GOOS == "windows" {
			pathSep = ";"
		}
		paths := strings.Split(kubeconfigEnv, pathSep)
		out := []string{}
		for _, myPath := range paths {
			if _, err := os.Stat(myPath); err == nil {
				out = append(out, myPath)
			}
		}
		return out, nil
	}

	// KUBECONFIG is not set.
	// Look for the default location ~/.kube/config
	filePath := path.Join(homeDir, ".kube", "config")

	if _, err := os.Stat(filePath); err == nil {
		// file exists in default location
		return []string{filePath}, nil
	}

	// No kubeconfig file. Return empty slice.
	return nil, nil
}

// GetDefaultCluster determines which is the default cluster. This can be either
// the only cluster accessible, or a cluster selected explicitly.
func GetDefaultCluster() (clusterID string, err error) {
	// Check selected cluster
	if Config.Cluster != "" {
		return Config.Cluster, nil
	}
	// Go through available orgs and clusters to find all clusters
	if Config.Token == "" {
		return "", errors.New("User not logged in.")
	}
	client := gsclientgen.NewDefaultApi()
	authHeader := "giantswarm " + Config.Token
	orgsResponse, _, err := client.GetUserOrganizations(authHeader)
	if err != nil {
		return "", err
	}
	if orgsResponse.StatusCode == 10000 {
		if len(orgsResponse.Data) > 0 {
			clusterIDs := []string{}
			for _, orgName := range orgsResponse.Data {
				clustersResponse, _, err := client.GetOrganizationClusters(authHeader, orgName)
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
