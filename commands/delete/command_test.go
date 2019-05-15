package delete

import "testing"

func TestCobraCommand(t *testing.T) {
	Command.SetArgs([]string{"--help"})
	Command.Execute()
}
