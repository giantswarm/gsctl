package commands

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/giantswarm/gsctl/config"
)

// create a temporary directory
func tempDir() string {
	dir, _ := ioutil.TempDir("", config.ProgramName)
	return dir
}

// tempConfig creates a temporary config directory with config.yaml file
// containing the given YAML content and initializes our config from it.
// The directory path ist returned.
func tempConfig(configYAML string) (string, error) {
	dir := tempDir()
	filePath := path.Join(dir, config.ConfigFileName+"."+config.ConfigFileType)

	if configYAML != "" {
		file, fileErr := os.Create(filePath)
		if fileErr != nil {
			return dir, fileErr
		}
		file.WriteString(configYAML)
		file.Close()
	}

	err := config.Initialize(dir)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

// create a temporary kubectl config file
func tempKubeconfig() (string, error) {

	// override standard paths for testing
	dir := tempDir()
	config.HomeDirPath = dir
	config.DefaultConfigDirPath = path.Join(config.HomeDirPath, ".config", config.ProgramName)

	// add a test kubectl config file
	kubeConfigPath := path.Join(dir, "tempkubeconfig")
	config.KubeConfigPaths = []string{kubeConfigPath}
	kubeConfig := []byte(`apiVersion: v1
kind: Config
preferences: {}
current-context: g8s-system
clusters:
users:
contexts:
`)
	fileErr := ioutil.WriteFile(kubeConfigPath, kubeConfig, 0700)
	if fileErr != nil {
		return "", fileErr
	}

	return kubeConfigPath, nil
}
