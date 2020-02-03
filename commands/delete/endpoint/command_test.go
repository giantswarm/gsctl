package endpoint

import (
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
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

// TestDeleteEndpointSuccess runs test case that are supposed to succeed
func TestDeleteEndpointSuccess(t *testing.T) {
	var testCases = []Arguments{
		{
			apiEndpoint: "foo",
			force:       true,
		},
		{
			apiEndpoint: "https://bar",
			force:       true,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, testCase := range testCases {

		args := testCase

		validateErr := validatePreconditions(args)
		if validateErr != nil {
			t.Errorf("Validation error in testCase %v: %s", i, validateErr.Error())
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
				apiEndpoint: "",
				force:       true,
			},
			expectedError: errors.EndpointMissingError,
		},
		{
			arguments: Arguments{
				apiEndpoint: "baz",
				force:       true,
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
