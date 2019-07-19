package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
)

// TestDeleteAppSuccess runs test case that are supposed to succeed
func TestDeleteAppSuccess(t *testing.T) {
	// mock server always responds positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": ""}`))
	}))
	defer mockServer.Close()

	var testCases = []deleteAppArguments{
		{
			apiEndpoint: mockServer.URL,
			clusterID:   "somecluster",
			appName:     "testapp",
			token:       "fake token",
			force:       true,
		},
	}

	flags.CmdAPIEndpoint = mockServer.URL

	for i, testCase := range testCases {

		flags.CmdToken = testCase.token
		flags.CmdForce = testCase.force

		args := defaultArguments([]string{testCase.appName})

		validateErr := validatePreconditions(args)
		if validateErr != nil {
			t.Errorf("Validation error in testCase %v: %s", i, validateErr.Error())
		} else {
			_, execErr := deleteApp(testCase)
			if execErr != nil {
				t.Errorf("Execution error in testCase %v: %s", i, execErr.Error())
			}
		}
	}
}

type failTestCase struct {
	arguments     deleteAppArguments
	expectedError error
}

// TestDeleteClusterFailures runs test case that are supposed to fail
func TestDeleteClusterFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		{
			arguments: deleteAppArguments{
				clusterID: "somecluster",
				appName:   "testapp",
				token:     "",
			},
			expectedError: errors.NotLoggedInError,
		},
		{
			arguments: deleteAppArguments{
				clusterID: "",
				appName:   "",
				token:     "some token",
			},
			expectedError: errors.ClusterIDMissingError,
		},
	}

	for i, ftc := range failTestCases {
		validateErr := validatePreconditions(ftc.arguments)
		if validateErr == nil {
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
	// mock server always responds positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": ""}`))
	}))
	defer mockServer.Close()

	testutils.CaptureOutput(func() {
		Command.SetArgs([]string{"--force", "--endpoint", mockServer.URL})
		Command.Execute()
	})
}
