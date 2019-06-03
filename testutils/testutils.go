package testutils

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/spf13/afero"

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
func TempDir(fs afero.Fs) string {
	dir, err := afero.TempDir(fs, "", config.ProgramName)
	if err != nil {
		panic(err)
	}
	return dir
}

// TempConfig creates a temporary config directory with config.yaml file
// containing the given YAML content and initializes our config from it.
// The directory path is returned.
func TempConfig(fs afero.Fs, configYAML string) (string, error) {
	dir := TempDir(fs)
	filePath := path.Join(dir, config.ConfigFileName+"."+config.ConfigFileType)

	if configYAML != "" {
		file, fileErr := fs.Create(filePath)
		if fileErr != nil {
			return dir, fileErr
		}
		file.WriteString(configYAML)
		file.Close()
	}

	err := config.Initialize(fs, dir)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

// TempKubeconfig creates a temporary kubectl config file for testing.
func TempKubeconfig(fs afero.Fs) (string, error) {
	// override standard paths for testing
	dir := TempDir(fs)
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
	err := afero.WriteFile(fs, kubeConfigPath, kubeConfig, 0700)
	if err != nil {
		return "", err
	}

	return kubeConfigPath, nil
}
