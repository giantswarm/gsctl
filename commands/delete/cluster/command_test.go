package cluster

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// TestDeleteClusterSuccess runs test case that are supposed to succeed
func TestDeleteClusterSuccess(t *testing.T) {
	// mock server always responds positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"code": "RESOURCE_DELETION_STARTED", "message": "We'll soon nuke this cluster"}`))
	}))
	defer mockServer.Close()

	var testCases = []Arguments{
		{
			apiEndpoint: mockServer.URL,
			clusterID:   "somecluster",
			token:       "fake-token",
			force:       true,
		},
	}

	flags.CmdAPIEndpoint = mockServer.URL

	for i, testCase := range testCases {

		flags.CmdToken = testCase.token
		flags.CmdForce = testCase.force
		args := testCase

		validateErr := validatePreconditions(args)
		if validateErr != nil {
			t.Errorf("Validation error in testCase %v: %s", i, validateErr.Error())
		} else {
			_, execErr := deleteCluster(testCase)
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

// TestDeleteClusterFailures runs test case that are supposed to fail
func TestDeleteClusterFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		{
			arguments: Arguments{
				clusterID: "somecluster",
				token:     "",
			},
			expectedError: errors.NotLoggedInError,
		},
		{
			arguments: Arguments{
				clusterID: "",
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
		w.Write([]byte(`{"code": "RESOURCE_DELETION_STARTED", "message": "We'll soon nuke this cluster"}`))
	}))
	defer mockServer.Close()

	testutils.CaptureOutput(func() {
		Command.SetArgs([]string{"--force", "--endpoint", mockServer.URL})
		Command.Execute()
	})
}
