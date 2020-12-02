package nodepool

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// configYAML is a mock configuration used by some of the tests.
const configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  %s:
    email: email@example.com
    token: some-token
selected_endpoint: %s
updated: 2017-09-29T11:23:15+02:00
`

// TestCollectArgs tests whether collectArguments produces the expected results.
func TestCollectArgs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.URL.String() == "/v4/info/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"general": {
				  "provider": "aws"
				},
				"features": {
				  "nodepools": {"release_version_minimum": "9.0.0"},
				  "ha_masters": {"release_version_minimum": "11.5.0"}
				}
			  }`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "ERROR", "message": "Bad things happened"}`))
		}
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, fmt.Sprintf(configYAML, mockServer.URL, mockServer.URL))
	if err != nil {
		t.Fatal(err)
	}

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
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				Provider:        "aws",
			},
		},
		{
			[]string{"clusterid/nodepoolid", "--nodes-min=3", "--nodes-max=5", "--name=NewName"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"clusterid/nodepoolid", "--nodes-min=3", "--nodes-max=5", "--name=NewName"})
			},
			Arguments{
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				ScalingMin:      3,
				ScalingMinSet:   true,
				ScalingMax:      5,
				Name:            "NewName",
				Provider:        "aws",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			Command.ParseFlags(tc.positionalArguments)

			tc.commandExecution()
			args, err := collectArguments(Command, tc.positionalArguments)
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
		name         string
		args         Arguments
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: cluster ID is missing",
			args: Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				Provider:    "aws",
				NodePoolID:  "abc",
			},
			errorMatcher: errors.IsClusterNameOrIDMissingError,
		},
		{
			name: "case 1: node pool ID is missing",
			args: Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Provider:        "aws",
				ClusterNameOrID: "abc",
			},
			errorMatcher: errors.IsNodePoolIDMissingError,
		},
		{
			name: "case 2: no token provided",
			args: Arguments{
				APIEndpoint:     "https://mock-url",
				Provider:        "aws",
				ClusterNameOrID: "cluster-id",
			},
			errorMatcher: errors.IsNotLoggedInError,
		},
		{
			name: "case 3: nothing to change, on aws",
			args: Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Provider:        "aws",
				ClusterNameOrID: "cluster-id",
				NodePoolID:      "abc",
			},
			errorMatcher: errors.IsNoOpError,
		},
		{
			name: "case 4: bad scaling parameters, on aws",
			args: Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Provider:        "aws",
				ClusterNameOrID: "cluster-id",
				NodePoolID:      "abc",
				ScalingMin:      10,
				ScalingMax:      1,
			},
			errorMatcher: errors.IsWorkersMinMaxInvalid,
		},
		{
			name: "case 5: nothing to change, on azure",
			args: Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Provider:        "azure",
				ClusterNameOrID: "cluster-id",
				NodePoolID:      "abc",
			},
			errorMatcher: errors.IsNoOpError,
		},
		{
			name: "case 6: trying to provide unsupported arguments, on azure",
			args: Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				Provider:        "azure",
				ClusterNameOrID: "cluster-id",
				NodePoolID:      "abc",
				ScalingMin:      1,
				ScalingMax:      3,
			},
			errorMatcher: errors.IsWorkersMinMaxInvalid,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				AuthToken:       "token",
				Name:            "New name",
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
						Min: testutils.Int64Value(3),
						Max: 5,
					},
				},
			},
		},
		// Scaling change.
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				AuthToken:       "token",
				ScalingMin:      10,
				ScalingMax:      20,
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
						Min: testutils.Int64Value(10),
						Max: 20,
					},
				},
			},
		},
		// Scaling change min only.
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				AuthToken:       "token",
				ScalingMin:      10,
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
						Min: testutils.Int64Value(10),
						Max: 20,
					},
				},
			},
		},
		// Scaling change max only.
		{
			Arguments{
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
				AuthToken:       "token",
				ScalingMax:      10,
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
						Min: testutils.Int64Value(3),
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
				} else if r.Method == "GET" && r.URL.Path == "/v5/clusters/clusterid/nodepools/nodepoolid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "nodepoolid",
						"name": "Some random name",
						"scaling": {
							"min": 999,
							"max": 1000
						}
					}`))
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
				ClusterNameOrID: "clusterid",
				NodePoolID:      "nodepoolid",
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
				NodePoolID:      "nodepoolid",
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
				NodePoolID:      "nodepoolid",
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
				NodePoolID:      "nodepoolid",
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
				if r.Method == "PATCH" && r.URL.Path == "/v5/clusters/clusterid/nodepools/nodepoolid/" {
					w.WriteHeader(tc.responseStatusCode)
					w.Write([]byte(tc.responseBody))
				} else if r.Method == "GET" && r.URL.Path == "/v5/clusters/clusterid/nodepools/nodepoolid/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "nodepoolid",
						"name": "Some random name",
						"scaling": {
							"min": 999,
							"max": 1000
						}
					}`))
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

			_, err := updateNodePool(tc.args)
			if err == nil {
				t.Errorf("Case %d - Expected error, got nil", i)
			} else if !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, err)
			}
		})
	}
}
