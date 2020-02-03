package endpoint

import (
	"strconv"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

const configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  https://foo:
    alias: foo
    email: email@example.com
    token: some-token
  https://bar:
    alias: bar
    email: email@example.com
    token: some-token
selected_endpoint: https://foo
updated: 2017-09-29T11:23:15+02:00
`

// TestCollectArgs tests whether collectArguments produces the expected results
func TestCollectArgs(t *testing.T) {
	var testCases = []struct {
		// The positional arguments we pass.
		positionalArguments []string
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{"https://foo"},
			Arguments{
				APIEndpoint: "https://foo",
				Force:       false,
				Verbose:     false,
			},
		},
		{
			[]string{"foo"},
			Arguments{
				APIEndpoint: "foo",
				Force:       false,
				Verbose:     false,
			},
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			args := collectArguments(tc.positionalArguments)
			if err != nil {
				t.Errorf("Case %d - Unexpected error '%s'", i, err)
			}
			if diff := cmp.Diff(tc.resultingArgs, args); diff != "" {
				t.Errorf("Case %d - Resulting args unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// TestDeleteEndpointSuccess runs test case that are supposed to succeed
func TestDeleteEndpointSuccess(t *testing.T) {
	var testCases = []Arguments{
		{
			APIEndpoint: "foo",
			Force:       true,
		},
		{
			APIEndpoint: "https://bar",
			Force:       true,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, testCase := range testCases {

		args := testCase

		err = validatePreconditions(args)
		if err != nil {
			t.Errorf("Validation error in testCase %v: %s", i, err.Error())
		} else {
			_, execErr := deleteEndpoint(args)
			if execErr != nil {
				t.Errorf("Execution error in testCase %v: %s", i, execErr.Error())
			}
		}
	}
}

type failTestCase struct {
	arguments     Arguments
	expectedError error
}

// TestDeleteEndpointFailures runs test case that are supposed to fail
func TestDeleteEndpointFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		{
			arguments: Arguments{
				APIEndpoint: "",
				Force:       true,
			},
			expectedError: errors.EndpointMissingError,
		},
		{
			arguments: Arguments{
				APIEndpoint: "baz",
				Force:       true,
			},
			expectedError: errors.EndpointNotFoundError,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, ftc := range failTestCases {
		_, err = deleteEndpoint(ftc.arguments)
		if err == nil {
			t.Errorf("Didn't get an error where we expected '%s' in testCase %v", ftc.expectedError, i)
		}
	}
}

func TestPrintResult(t *testing.T) {
	var testCases = []struct {
		apiEndpoint   string
		outputMessage string
	}{
		{
			apiEndpoint:   "baz",
			outputMessage: "API Endpoint not found\nThe API endpoint you are trying to delete does not exist. Check 'gsctl list endpoints' to make sure",
		},
		{
			apiEndpoint:   "https://bar",
			outputMessage: "The API endpoint 'https://bar' deleted successfully.",
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, testCase := range testCases {
		output := testutils.CaptureOutput(func() {
			initFlags()
			Command.ParseFlags([]string{"--force"})
			printResult(Command, []string{testCase.apiEndpoint})
		})

		if !strings.Contains(output, testCase.outputMessage) {
			t.Errorf("Case %d: Missing '%s' from output", i, testCase.outputMessage)
		}
	}
}

func TestCommandExecutionHelp(t *testing.T) {
	testutils.CaptureOutput(func() {
		Command.SetArgs([]string{"--help"})
		Command.Execute()
	})
}

func TestCommandExecution(t *testing.T) {
	testutils.CaptureOutput(func() {
		Command.SetArgs([]string{"--force"})
		Command.Execute()
	})
}
