package commands

import (
	"io/ioutil"
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

func Test_PrintInfoWithTempDirAndToken(t *testing.T) {
	dir, _ := ioutil.TempDir("", config.ProgramName)
	cmdConfigDirPath = dir
	cmdToken = "fake token"
	printInfo(InfoCommand, []string{})
}
