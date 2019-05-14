package info

import (
	"os"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/flags"
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

	dir, err := testutils.TempConfig(yamlText)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	output := testutils.CaptureOutput(func() {
		Command.Execute()
	})
	if strings.Contains(output, "Auth token:") {
		t.Error("Verbose Command output did not contain 'Auth token'")
	}
}

// Test_PrintInfoVerbose simply executes the printInfo function with verbose=true
// TODO: actually test what this does
func Test_PrintInfoVerbose(t *testing.T) {
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

	dir, err := testutils.TempConfig(yamlText)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	flags.CmdVerbose = true

	output := testutils.CaptureOutput(func() {
		Command.Execute()
	})
	if !strings.Contains(output, "Auth token:") {
		t.Error("Verbose Command output did not contain 'Auth token'")
	}
}

// Test_InfoWithTempDirAndToken tests the info() function with a custom
// configuration path and an auth-token
func Test_InfoWithTempDirAndToken(t *testing.T) {
	dir, err := testutils.TempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	flags.CmdAPIEndpoint = ""
	args := defaultInfoArguments()
	args.token = "fake token"

	infoResult, err := info(args)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(infoResult.configFilePath, dir) {
		t.Errorf("Config file path not as expected: Got %s, expected %s",
			infoResult.configFilePath, dir)
	}
	if infoResult.token != args.token {
		t.Errorf("Expected token '%s', got '%s'", flags.CmdToken, infoResult.token)
	}
	if infoResult.email != "" {
		t.Error("Expected empty email, got ", infoResult.email)
	}
}
