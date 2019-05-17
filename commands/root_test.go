package commands

import (
	"testing"
)

func Test_RootCommand(t *testing.T) {
	err := RootCommand.Execute()
	if err != nil {
		t.Error(err)
	}
}
