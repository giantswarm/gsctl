package main

import (
	"testing"

	"github.com/giantswarm/gsctl/commands"
)

func Test_Main(t *testing.T) {
	err := commands.RootCommand.Execute()
	if err != nil {
		t.Error(err)
	}
}
