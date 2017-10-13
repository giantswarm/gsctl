package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// Test_ListEndpoints tests the listing of a few endpoints
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

	args := listEndpointsArguments{
		// "" here means that we don't override the selected endpoint
		apiEndpoint: config.Config.ChooseEndpoint(""),
	}

	table := endpointsTable(args)
	if table == "" {
		t.Error("Got no output where I expected a table")
	}

	testString := "https://my.second.endpoint  email@example.com  yes       yes"
	if !strings.Contains(table, testString) {
		t.Errorf("Table does not contain expected row '%s'", testString)
	}
}
