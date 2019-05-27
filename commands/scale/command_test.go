package scale

import "testing"

func TestCobraCommand(t *testing.T) {
	Command.SetArgs([]string{"--help"})
	err := Command.Execute()
	if err != nil {
		t.Error(err)
	}
}
