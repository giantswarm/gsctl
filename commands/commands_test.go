package commands

import (
	"io/ioutil"
	"path"

	"github.com/giantswarm/gsctl/config"
)

// create a temporary kubectl config file
func tempKubeconfig() (string, error) {

	// override standard paths for testing
	dir := config.TempDir()
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
