package config

import (
	"io/ioutil"
	"os"
	"path"
)

// TempDir creates a temporary directory for a temporary config file in tests.
func TempDir() string {
	dir, err := ioutil.TempDir("", ProgramName)
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
	filePath := path.Join(dir, ConfigFileName+"."+ConfigFileType)

	if configYAML != "" {
		file, fileErr := os.Create(filePath)
		if fileErr != nil {
			return dir, fileErr
		}
		file.WriteString(configYAML)
		file.Close()
	}

	err := Initialize(dir)
	if err != nil {
		return dir, err
	}

	return dir, nil
}
