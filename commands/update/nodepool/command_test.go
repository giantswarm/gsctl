package nodepool

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

// configYAML is a mock configuration used by some of the tests.
const configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  https://foo:
    email: email@example.com
    token: some-token
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
			[]string{"clusterid/nodepoolid"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"clusterid/nodepoolid"})
			},
			Arguments{
				APIEndpoint: "https://foo",
				AuthToken:   "some-token",
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
			},
		},
		{
			[]string{"clusterid/nodepoolid", "--nodes-min=3", "--nodes-max=5", "--name=NewName"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"clusterid/nodepoolid", "--nodes-min=3", "--nodes-max=5", "--name=NewName"})
			},
			Arguments{
				APIEndpoint: "https://foo",
				AuthToken:   "some-token",
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
				ScalingMin:  3,
				ScalingMax:  5,
				Name:        "NewName",
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
			args, err := collectArguments(tc.positionalArguments)
			if err != nil {
				t.Errorf("Case %d - Unexpected error '%s'", i, err)
			}
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
				NodePoolID:  "abc",
			},
			errors.IsClusterIDMissingError,
		},
		// Node pool ID is missing.
		{
			Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				ClusterID:   "abc",
			},
			errors.IsNodePoolIDMissingError,
		},
		// No token provided.
		{
			Arguments{
				APIEndpoint: "https://mock-url",
				ClusterID:   "cluster-id",
			},
			errors.IsNotLoggedInError,
		},
		// Nothing to change.
		{
			Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				ClusterID:   "cluster-id",
				NodePoolID:  "abc",
			},
			errors.IsNoOpError,
		},
		// Bad scaling parameters
		{
			Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				ClusterID:   "cluster-id",
				NodePoolID:  "abc",
				ScalingMin:  10,
				ScalingMax:  1,
			},
			errors.IsWorkersMinMaxInvalid,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := verifyPreconditions(tc.args)
			if err == nil {
				t.Errorf("Case %d - Expected error, got nil", i)
			} else if !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, err)
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
			Arguments{
				ClusterID:  "clusterid",
				NodePoolID: "nodepoolid",
				AuthToken:  "token",
				Name:       "New name",
			},
			`{
				"id": "nodepoolid",
				"name": "New name",
				"scaling": {
					"min": 3,
					"max": 5
				}
			}`,
			&result{
				NodePool: &models.V5GetNodePoolResponse{
					ID:   "nodepoolid",
					Name: "New name",
					Scaling: &models.V5GetNodePoolResponseScaling{
						Min: 3,
						Max: 5,
					},
				},
			},
		},
		// Scaling change.
		{
			Arguments{
				ClusterID:  "clusterid",
				NodePoolID: "nodepoolid",
				AuthToken:  "token",
				ScalingMin: 10,
				ScalingMax: 20,
			},
			`{
				"id": "nodepoolid",
				"name": "New name",
				"scaling": {
					"min": 10,
					"max": 20
				}
			}`,
			&result{
				NodePool: &models.V5GetNodePoolResponse{
					ID:   "nodepoolid",
					Name: "New name",
					Scaling: &models.V5GetNodePoolResponseScaling{
						Min: 10,
						Max: 20,
					},
				},
			},
		},
		// Scaling change min only.
		{
			Arguments{
				ClusterID:  "clusterid",
				NodePoolID: "nodepoolid",
				AuthToken:  "token",
				ScalingMin: 10,
			},
			`{
				"id": "nodepoolid",
				"name": "New name",
				"scaling": {
					"min": 10,
					"max": 20
				}
			}`,
			&result{
				NodePool: &models.V5GetNodePoolResponse{
					ID:   "nodepoolid",
					Name: "New name",
					Scaling: &models.V5GetNodePoolResponseScaling{
						Min: 10,
						Max: 20,
					},
				},
			},
		},
		// Scaling change max only.
		{
			Arguments{
				ClusterID:  "clusterid",
				NodePoolID: "nodepoolid",
				AuthToken:  "token",
				ScalingMax: 10,
			},
			`{
				"id": "nodepoolid",
				"name": "New name",
				"scaling": {
					"min": 3,
					"max": 10
				}
			}`,
			&result{
				NodePool: &models.V5GetNodePoolResponse{
					ID:   "nodepoolid",
					Name: "New name",
					Scaling: &models.V5GetNodePoolResponseScaling{
						Min: 3,
						Max: 10,
					},
				},
			},
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// set up mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Logf("Mock request: %s %s", r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "PATCH" && r.URL.Path == "/v5/clusters/clusterid/nodepools/nodepoolid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tc.responseBody))
				} else {
					t.Errorf("Case %d - Unsupported operation %s %s called in mock server", i, r.Method, r.URL.Path)
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			err := verifyPreconditions(tc.args)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			result, err := updateNodePool(tc.args)
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
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				Name:        "weird-name",
			},
			401,
			`{"code": "FORBIDDEN", "message": "Here is some error message"}`,
			errors.IsNotAuthorizedError,
		},
		{
			Arguments{
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				Name:        "weird-name",
			},
			403,
			`{"code": "FORBIDDEN", "message": "Here is some error message"}`,
			errors.IsAccessForbiddenError,
		},
		{
			Arguments{
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				Name:        "weird-name",
			},
			404,
			`{"code": "RESOURCE_NOT_FOUND", "message": "Here is some error message"}`,
			errors.IsClusterNotFoundError,
		},
		{
			Arguments{
				ClusterID:   "clusterid",
				NodePoolID:  "nodepoolid",
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				Name:        "weird-name",
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
				if r.Method == "PATCH" && r.URL.Path == "/v5/clusters/clusterid/nodepools/nodepoolid/" {
					w.WriteHeader(tc.responseStatusCode)
					w.Write([]byte(tc.responseBody))
				} else {
					t.Errorf("Unsupported operation %s %s called in mock server", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			_, err := updateNodePool(tc.args)
			if err == nil {
				t.Errorf("Case %d - Expected error, got nil", i)
			} else if !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, err)
			}
		})
	}
}
