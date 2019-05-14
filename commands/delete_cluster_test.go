package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
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

	var testCases = []deleteClusterArguments{
		{
			apiEndpoint: mockServer.URL,
			clusterID:   "somecluster",
			token:       "fake token",
			force:       true,
		},
	}

	flags.CmdAPIEndpoint = mockServer.URL
	InitClient()

	for i, testCase := range testCases {
		validateErr := validateDeleteClusterPreConditions(testCase)
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
	arguments     deleteClusterArguments
	expectedError error
}

// TestDeleteClusterFailures runs test case that are supposed to fail
func TestDeleteClusterFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		{
			arguments: deleteClusterArguments{
				clusterID: "somecluster",
				token:     "",
			},
			expectedError: errors.NotLoggedInError,
		},
		{
			arguments: deleteClusterArguments{
				clusterID: "",
				token:     "some token",
			},
			expectedError: errors.ClusterIDMissingError,
		},
	}

	for i, ftc := range failTestCases {
		validateErr := validateDeleteClusterPreConditions(ftc.arguments)
		if validateErr == nil {
			t.Errorf("Didn't get an error where we expected '%s' in testCase %v", ftc.expectedError, i)
		}
	}
}
