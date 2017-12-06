package commands

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// create a temporary directory
func tempDir() string {
	dir, err := ioutil.TempDir("", config.ProgramName)
	if err != nil {
		panic(err)
	}
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

// TestRedactPasswordArgs tests redactPasswordArgs()
func TestRedactPasswordArgs(t *testing.T) {
	argtests := []struct {
		in  string
		out string
	}{
		// these remain unchangd
		{"foo", "foo"},
		{"foo bar", "foo bar"},
		{"foo bar blah", "foo bar blah"},
		{"foo bar blah -p mypass", "foo bar blah -p mypass"},
		{"foo bar blah -p=mypass", "foo bar blah -p=mypass"},
		// these will be altered
		{"foo bar blah --password mypass", "foo bar blah --password REDACTED"},
		{"foo bar blah --password=mypass", "foo bar blah --password=REDACTED"},
		{"foo login blah -p mypass", "foo login blah -p REDACTED"},
		{"foo login blah -p=mypass", "foo login blah -p=REDACTED"},
	}

	for _, tt := range argtests {
		in := strings.Split(tt.in, " ")
		out := strings.Join(redactPasswordArgs(in), " ")
		if out != tt.out {
			t.Errorf("want '%q', have '%s'", tt.in, tt.out)
		}
	}
}
