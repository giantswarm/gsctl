package info

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/testutils"
)

// Test_PrintInfo simply executes the printInfo function.
// TODO: actually test what this does
func Test_PrintInfo(t *testing.T) {
	// our test config YAML
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  https://myapi.domain.tld:
    email: email@example.com
    token: some-token
    alias: myalias
  https://other.endpoint:
    email: ""
    token: ""
    alias: ""
selected_endpoint: https://other.endpoint`

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
	if err != nil {
		t.Error(err)
	}

	output := testutils.CaptureOutput(func() {
		Command.Execute()
	})
	t.Log(output)

	if strings.Contains(output, "Auth token:") {
		t.Error("Verbose Command output did not contain 'Auth token'")
	}

	re := regexp.MustCompile(`Email:\s+n/a`)
	if re.Find([]byte(output)) == nil {
		t.Error("Output did not contain expected chunk 'Email: n/a'")
	}
}

// Test_InfoWithTempDirAndToken tests the info() function with a custom
// configuration path and an auth-token
func Test_InfoWithTempDirAndToken(t *testing.T) {
	fs := afero.NewMemMapFs()
	dir, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	args := collectArguments()
	args.token = "fake token"
	args.userProvidedToken = args.token
	args.apiEndpoint = ""

	infoResult, err := info(args)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(infoResult.configFilePath, dir) {
		t.Errorf("Config file path not as expected: Got %s, expected %s",
			infoResult.configFilePath, dir)
	}
	if infoResult.token != args.token {
		t.Errorf("Expected token '%s', got '%s'", args.token, infoResult.token)
	}
	if infoResult.email != "" {
		t.Error("Expected empty email, got ", infoResult.email)
	}
}
