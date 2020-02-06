package util

import (
	"testing"

	"github.com/spf13/cobra"
)

var testCommand = &cobra.Command{
	Use: "test-command",
}

var getCompletionFnTests = []struct {
	fName  string
	fBody  string
	output string
}{
	{
		"test",
		"ls",
		`
test() {
	ls
}`,
	},
}

func TestGetBashCompletionFn(t *testing.T) {
	for i, test := range getCompletionFnTests {
		out := GetBashCompletionFn(test.fName, test.fBody)

		if out != test.output {
			t.Errorf("#%d: TestGetBashCompletionFn(%s, %s) = '%s'; want '%s'", i, test.fName, test.fBody, out, test.output)
		}
	}
}

var setFlagCompletionFnTests = []struct {
	fName  string
	fBody  string
	output string
}{
	{
		"test",
		"ls",
		`
test() {
	ls
}`,
	},
	{
		"test2",
		"ls",
		`
test() {
	ls
}
test2() {
	ls
}`,
	},
	{
		"test",
		"ls",
		`
test() {
	ls
}
test2() {
	ls
}`,
	},
}

func TestSetFlagBashCompletionFn(t *testing.T) {
	for i, test := range setFlagCompletionFnTests {
		SetFlagBashCompletionFn(&BashCompletionFunc{
			Command:  testCommand,
			Flags:    testCommand.PersistentFlags(),
			FlagName: "test-flag",
			FnName:   test.fName,
			FnBody:   test.fBody,
		})
		out := testCommand.Root().BashCompletionFunction

		if out != test.output {
			t.Errorf("#%d: SetFlagBashCompletionFn(%s, %s) = '%s'; want '%s'", i, test.fName, test.fBody, out, test.output)
		}
	}

	testCommand.Root().BashCompletionFunction = ""
}

var customCommandCompletionTests = []struct {
	fBody  string
	output string
}{
	{
		"ls",
		`ls
`,
	},
	{
		"pwd",
		`ls
pwd
`,
	},
}

func TestCustomCommandCompletion(t *testing.T) {
	for i, test := range customCommandCompletionTests {
		SetCommandBashCompletion(&BashCompletionFunc{
			FnBody: test.fBody,
		})
		out := GetCustomCommandCompletionFnBody()

		if out != test.output {
			t.Errorf("#%d: TestCustomCommandCompletion(%s) = '%s'; want '%s'", i, test.fBody, out, test.output)
		}
	}
}
