package cluster

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/spf13/afero"
)

// TestDeleteClusterSuccess runs test case that are supposed to succeed
func TestDeleteClusterSuccess(t *testing.T) {
	// mock server always responds positively
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("mockServer request: ", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" && r.URL.String() == "/v4/clusters/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
			{
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"id": "as123asd",
				"name": "somecluster",
				"owner": "acme",
				"path": "/v4/clusters/as123asd/"
			},
			{
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"id": "someothercluster",
				"name": "My dearest production cluster",
				"owner": "acme",
				"path": "/v4/clusters/somecluster/"
			}
		]`))
		} else {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"code": "RESOURCE_DELETION_STARTED", "message": "We'll soon nuke this cluster"}`))
		}
	}))
	defer mockServer.Close()

	var testCases = []Arguments{
		{
			apiEndpoint:     mockServer.URL,
			clusterNameOrID: "somecluster",
			token:           "fake-token",
			force:           true,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	for i, testCase := range testCases {

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
	arguments    Arguments
	errorMatcher func(error) bool
}

// TestDeleteClusterFailures runs test case that are supposed to fail
func TestDeleteClusterFailures(t *testing.T) {
	var failTestCases = []failTestCase{
		{
			arguments: Arguments{
				apiEndpoint:     "https://mock-url",
				clusterNameOrID: "somecluster",
				token:           "",
			},
			errorMatcher: errors.IsNotLoggedInError,
		},
		{
			arguments: Arguments{
				apiEndpoint:     "https://mock-url",
				clusterNameOrID: "",
				token:           "some token",
			},
			errorMatcher: errors.IsClusterNameOrIDMissingError,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	for i, ftc := range failTestCases {
		validateErr := validatePreconditions(ftc.arguments)
		if validateErr == nil {
			t.Errorf("Case %d - Expected error, got nil", i)
		} else if !ftc.errorMatcher(validateErr) {
			t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, validateErr)
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
