package cluster

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// configYAML is a mock configuration used by some of the tests.
const configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  https://foo:
    email: email@example.com
    token: some-token
    provider: aws
selected_endpoint: https://foo
updated: 2017-09-29T11:23:15+02:00
`

// TestCollectArgs tests whether collectArguments produces the expected results.
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
			[]string{"clusterid"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"clusterid"})
			},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
			},
		},
		{
			[]string{"clusterid", "--name=NewName"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"clusterid", "--name=NewName"})
			},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Name:            "NewName",
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
			tc.commandExecution()
			args := collectArguments(tc.positionalArguments)

			if diff := cmp.Diff(tc.resultingArgs, args); diff != "" {
				t.Errorf("Case %d - Resulting args unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// Test_verifyPreconditions tests cases where validating preconditions fails.
func Test_verifyPreconditions(t *testing.T) {
	var testCases = []struct {
		args         Arguments
		errorMatcher func(error) bool
	}{
		// Cluster ID is missing.
		{
			Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
			},
			errors.IsClusterNameOrIDMissingError,
		},
		// No token provided.
		{
			Arguments{
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
			},
			errors.IsNotLoggedInError,
		},
		// Nothing to change.
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
			},
			nil,
		},
		// name and label arguments given at same time
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				Name:            "newname",
				Labels:          []string{"labelchange=one", "labelchange=two"},
			},
			errors.IsConflictingFlagsError,
		},
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				Name:            "newname",
				MasterHA:        true,
			},
			nil,
		},
		// HA Master has it's default value.
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				MasterHA:        false,
			},
			nil,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := verifyPreconditions(Command, tc.args)
			if tc.errorMatcher == nil {
				if err != nil {
					t.Errorf("Case %d - Error did not match expected type. Got '%s'", i, err)
				}
			} else {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Error did not match expected type. Got '%s'", i, err)
				}
			}
		})
	}
}

// TestSuccess tests node pool creation with cases that are expected to succeed.
func TestSuccess(t *testing.T) {
	var testCases = []struct {
		args           Arguments
		responseBody   string
		expectedResult *result
	}{
		// Name change.
		{
			args: Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				Name:            "New name",
			},
			responseBody: `{
				"id": "clusterid",
				"name": "New name"
			}`,
			expectedResult: &result{
				ClusterName: "New name",
			},
		},
		// Switch to HA masters.
		{
			args: Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				MasterHA:        true,
			},
			responseBody: `{
				"id": "clusterid",
				"master_nodes": {
                    "availability_zones": ["a", "b", "c"],
                    "high_availability": true,
                    "num_ready": 1
                }
			}`,
			expectedResult: &result{
				HasHAMaster: true,
			},
		},
		// Change name and switch to HA masters.
		{
			args: Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				Name:            "New name",
				MasterHA:        true,
			},
			responseBody: `{
				"id": "clusterid",
                "name": "New name",
				"master_nodes": {
                    "availability_zones": ["a", "b", "c"],
                    "high_availability": true,
                    "num_ready": 1
                }
			}`,
			expectedResult: &result{
				ClusterName: "New name",
				HasHAMaster: true,
			},
		},
		// Label change
		{
			args: Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				Labels: []string{
					"newlabelkey=newlabelvalue",
				},
			},
			responseBody: `{
				"labels": {
					"newlabelkey": "newlabelvalue"
				}
			}`,
			expectedResult: &result{
				Labels: map[string]string{
					"newlabelkey": "newlabelvalue",
				},
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
			// set up mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Logf("Mock request: %s %s", r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "PATCH" && r.URL.Path == "/v5/clusters/clusterid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tc.responseBody))
				} else if r.Method == "GET" && r.URL.Path == "/v5/clusters/clusterid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id": "clusterid", "name": "old cluster name", "release_version": "11.5.0"}`))
				} else if r.Method == "GET" && r.URL.Path == "/v4/clusters/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
						{
							"id": "clusterid",
							"name": "Name of the cluster",
							"owner": "acme",
							"release_version": "11.5.0"
						}
					]`))
				} else if r.Method == "PUT" && r.URL.Path == "/v5/clusters/clusterid/labels/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tc.responseBody))
				} else {
					t.Errorf("Case %d - Unsupported operation %s %s called in mock server", i, r.Method, r.URL.Path)
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			err := verifyPreconditions(Command, tc.args)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			result, err := updateCluster(tc.args)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			if diff := cmp.Diff(tc.expectedResult, result); diff != "" {
				t.Errorf("Case %d - Results unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// TestExecuteWithError tests the error handling.
func TestExecuteWithError(t *testing.T) {
	var testCases = []struct {
		args               Arguments
		responseStatusCode int
		responseBody       string
		errorMatcher       func(error) bool
	}{
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Name:            "weird-name",
			},
			401,
			`{"code": "FORBIDDEN", "message": "Here is some error message"}`,
			errors.IsNotAuthorizedError,
		},
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Name:            "weird-name",
			},
			403,
			`{"code": "FORBIDDEN", "message": "Here is some error message"}`,
			errors.IsAccessForbiddenError,
		},
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Name:            "weird-name",
			},
			404,
			`{"code": "RESOURCE_NOT_FOUND", "message": "Here is some error message"}`,
			errors.IsClusterNotFoundError,
		},
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Name:            "weird-name",
			},
			500,
			`{"code": "UNKNOWN_ERROR", "message": "Here is some error message"}`,
			errors.IsInternalServerError,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "PATCH" && r.URL.Path == "/v5/clusters/clusterid/" {
					w.WriteHeader(tc.responseStatusCode)
					w.Write([]byte(tc.responseBody))
				} else if r.Method == "GET" && r.URL.Path == "/v5/clusters/clusterid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id": "clusterid", "name": "old cluster name"}`))
				} else if r.Method == "GET" && r.URL.Path == "/v4/clusters/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
						{
							"id": "clusterid",
							"name": "Name of the cluster",
							"owner": "acme"
						}
					]`))
				} else {
					t.Errorf("Unsupported operation %s %s called in mock server", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			_, err := updateCluster(tc.args)
			if err == nil {
				t.Errorf("Case %d - Expected error, got nil", i)
			} else if !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, err)
			}
		})
	}
}

func Test_modifyClusterLabelsRequestFromArguments(t *testing.T) {
	mockLabels := []string{"this=works", "workstoo="}

	request, err := modifyClusterLabelsRequestFromArguments(mockLabels)
	if err != nil {
		t.Error(err)
	}

	if len(request.Labels) != 2 {
		t.Errorf("Invalid labels map length. Expected %d, got %d", 2, len(request.Labels))
	}

	mockLabels = []string{"missingequalsign", "perfectly=valid"}

	request, err = modifyClusterLabelsRequestFromArguments(mockLabels)
	if err == nil {
		t.Errorf("Expected error")
	}

	expectedErrorStr := "no op error: malformed label change 'missingequalsign' (single = required)"

	if err.Error() != expectedErrorStr {
		t.Errorf("Expected error to be '%s' got '%s'", expectedErrorStr, err.Error())
	}

	mockLabels = []string{"=invalid"}

	request, err = modifyClusterLabelsRequestFromArguments(mockLabels)
	if err == nil {
		t.Errorf("Expected error")
	}

	expectedErrorStr = "no op error: malformed label change '=invalid' (empty key)"

	if err.Error() != expectedErrorStr {
		t.Errorf("Expected error to be '%s' got '%s'", expectedErrorStr, err.Error())
	}

}
