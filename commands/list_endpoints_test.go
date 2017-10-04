package commands

import (
	"os"
	"path"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

func Test_ListEndpoints(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	filePath := path.Join(dir, config.ConfigFileName+"."+config.ConfigFileType)

	// dummy config
	file, fileErr := os.Create(filePath)
	if fileErr != nil {
		t.Error(fileErr)
	}

	// our test config YAML
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  https://my.first.endpoint:
    email: email@example.com
    token: some-token
  https://my.second.endpoint:
    email: email@example.com
    token: some-other-token
selected_endpoint: https://my.second.endpoint
`

	file.WriteString(yamlText)
	file.Close()

	err := config.Initialize(dir)
	if err != nil {
		t.Error("Error in Initialize:", err)
	}

	table := endpointsTable()
	if table == "" {
		t.Error("Got no output where I expected a table")
	}
}
