package testutils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/giantswarm/gsctl/config"
)

// CaptureOutput runs a function and captures returns STDOUT output as a string.
func CaptureOutput(f func()) (printed string) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = orig // restoring the real stdout
	out := <-outC

	return out
}

// TempDir creates a temporary directory for a temporary config file in tests.
func TempDir() string {
	dir, err := ioutil.TempDir("", config.ProgramName)
	if err != nil {
		panic(err)
	}
	return dir
}

// TempConfig creates a temporary config directory with config.yaml file
// containing the given YAML content and initializes our config from it.
// The directory path is returned.
func TempConfig(configYAML string) (string, error) {
	dir := TempDir()
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

// TempKubeconfig creates a temporary kubectl config file for testing.
func TempKubeconfig() (string, error) {
	// override standard paths for testing
	dir := TempDir()
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
