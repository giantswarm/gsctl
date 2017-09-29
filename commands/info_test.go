package commands

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

func Test_PrintInfo(t *testing.T) {
	printInfo(InfoCommand, []string{})
}

func Test_PrintInfoVerbose(t *testing.T) {
	cmdVerbose = true
	printInfo(InfoCommand, []string{})
}

// Test_InfoWithTempDirAndToken tests the info() function with a custom
// configuration path and an auth-token
func Test_InfoWithTempDirAndToken(t *testing.T) {
	dir, _ := ioutil.TempDir("", config.ProgramName)
	defer os.RemoveAll(dir)

	// Normally cobra does this for us, but here we don't use cobra.
	config.Initialize(dir)

	args := defaultInfoArguments()
	args.token = "fake token"

	infoResult := info(args)

	if !strings.Contains(infoResult.configFilePath, dir) {
		t.Errorf("Config file path not as expected: Got %s, expected %s",
			infoResult.configFilePath, dir)
	}
	if infoResult.token != args.token {
		t.Errorf("Expected token '%s', got '%s'", cmdToken, infoResult.token)
	}
	if infoResult.email != "" {
		t.Error("Expected empty email, got ", infoResult.email)
	}
}
