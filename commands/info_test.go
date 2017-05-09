package commands

import "testing"

func Test_Info(t *testing.T) {
	printInfo(InfoCommand, []string{})
}

func Test_Info_Verbose(t *testing.T) {
	cmdVerbose = true
	printInfo(InfoCommand, []string{})
}
