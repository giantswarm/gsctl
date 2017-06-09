package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteCluster(t *testing.T) {
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
