package endpoint

import (
	"strconv"
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
		// How we execute the command.
		commandExecution func()
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{"https://foo"},
			func() {
			},
			Arguments{
				APIEndpoint: "https://foo",
				Force:       false,
				Verbose:     false,
			},
		},
		{
			[]string{"foo"},
			func() {
			},
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
			// tc.commandExecution()
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

func TestPreconditionValidation(t *testing.T) {
	var testCases = []failTestCase{
		{
			arguments: Arguments{
				APIEndpoint: "",
				Force:       true,
				Verbose:     false,
			},
			expectedError: errors.EndpointMissingError,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, ftc := range testCases {
		err = validatePreconditions(ftc.arguments)
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
		Command.SetArgs([]string{"--Force"})
		Command.Execute()
	})
}
