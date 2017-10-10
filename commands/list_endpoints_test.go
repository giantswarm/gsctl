package commands

import (
	"os"
	"testing"
)

func Test_ListEndpoints(t *testing.T) {
	// dummy config
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
	dir, err := tempConfig(yamlText)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	table := endpointsTable()
	if table == "" {
		t.Error("Got no output where I expected a table")
	}
}
