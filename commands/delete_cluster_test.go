package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDeleteClusterSuccess runs test case that are supposed to succeed
func TestDeleteClusterSuccess(t *testing.T) {
	// mock server always responding positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETION_STARTED", "message": "We'll soon nuke this cluster"}`))
	}))
	defer mockServer.Close()

	var testCases = []deleteClusterArguments{
		deleteClusterArguments{
			apiEndpoint: mockServer.URL,
			clusterID:   "somecluster",
			token:       "fake token",
			force:       true,
		},
	}

	for i, testCase := range testCases {
		validateErr := validateDeleteClusterPreConditions(testCase)
		if validateErr != nil {
			t.Error(fmt.Sprintf("Validation error in testCase %v: %s", i, validateErr.Error()))
		} else {
			_, execErr := deleteCluster(testCase)
			if execErr != nil {
				t.Error(fmt.Sprintf("Execution error in testCase %v: %s", i, execErr.Error()))
			}
		}
	}
}

type failTestCase struct {
	arguments           deleteClusterArguments
	expectedErrorString string
}

// TestDeleteClusterFailures runs test case that are supposed to fail
func TestDeleteClusterFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		failTestCase{
			arguments: deleteClusterArguments{
				clusterID: "somecluster",
				token:     "",
			},
			expectedErrorString: errNotLoggedIn,
		},
		failTestCase{
			arguments: deleteClusterArguments{
				clusterID: "",
				token:     "some token",
			},
			expectedErrorString: errClusterIDNotSpecified,
		},
	}

	for i, ftc := range failTestCases {
		validateErr := validateDeleteClusterPreConditions(ftc.arguments)
		if validateErr == nil {
			t.Errorf("Didn't get an error where we expected '%s' in testCase %v", ftc.expectedErrorString, i)
		}
	}
}
