package commands

import (
	"os"
	"strings"
	"testing"
)

// Test_PrintInfo simply executes the printInfo function.
// TODO: actually test what this does
func Test_PrintInfo(t *testing.T) {
	printInfo(InfoCommand, []string{})
}

// Test_PrintInfoVerbose simply executes the printInfo function with verbose=true
// TODO: actually test what this does
func Test_PrintInfoVerbose(t *testing.T) {
	cmdVerbose = true
	printInfo(InfoCommand, []string{})
}

// Test_InfoWithTempDirAndToken tests the info() function with a custom
// configuration path and an auth-token
func Test_InfoWithTempDirAndToken(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

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
