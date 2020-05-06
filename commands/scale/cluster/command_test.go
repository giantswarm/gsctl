package cluster

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/client"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// configYAML is a mock configuration.
const configYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  https://foo:
    email: email@example.com
    token: some-token
selected_endpoint: https://foo
updated: 2017-09-29T11:23:15+02:00
`

func TestCollectArguments(t *testing.T) {
	var testCases = []struct {
		// The command line arguments passed
		cmdLineArgs []string
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{"clusterid"},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Scheme:          "giantswarm",
			},
		},
		{
			[]string{"clusterid", "--num-workers=5"},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Scheme:          "giantswarm",
				WorkersMax:      5,
				WorkersMaxSet:   false,
				WorkersMin:      5,
				WorkersMinSet:   false,
				Workers:         5,
				WorkersSet:      true,
			},
		},
		{
			[]string{"clusterid", "--workers-min=12"},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Scheme:          "giantswarm",
				WorkersMaxSet:   false,
				WorkersMin:      12,
				WorkersMinSet:   true,
				WorkersSet:      false,
			},
		},
		{
			[]string{"clusterid", "--workers-max=12"},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Scheme:          "giantswarm",
				WorkersMaxSet:   true,
				WorkersMax:      12,
				WorkersMinSet:   false,
				WorkersSet:      false,
			},
		},
		{
			[]string{"clusterid", "--num-workers=5", "--workers-min=4", "--workers-max=6"},
			Arguments{
				APIEndpoint:     "https://foo",
				AuthToken:       "some-token",
				ClusterNameOrID: "clusterid",
				Scheme:          "giantswarm",
				WorkersMax:      5,
				WorkersMaxSet:   true,
				WorkersMin:      5,
				WorkersMinSet:   true,
				Workers:         5,
				WorkersSet:      true,
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
			initFlags()
			Command.ParseFlags(tc.cmdLineArgs)

			args, err := collectArguments(Command, tc.cmdLineArgs)
			if err != nil {
				t.Errorf("Case %d - Unexpected error '%s'", i, err)
			}
			if diff := cmp.Diff(tc.resultingArgs, args); diff != "" {
				t.Errorf("Case %d - Resulting args unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

func TestVerifyPreconditions(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("mockServer: %s %s", r.Method, r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.URL.String() == "/v5/clusters/v5-cluster-id/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "v5-cluster-id",
				"name": "v5 cluster",
				"api_endpoint": "https://some-url",
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"owner": "acmeorg"
			}`))
		} else if r.Method == "GET" && r.URL.String() == "/v5/clusters/v5-cluster-id/nodepools/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Could not find this."}`))
		}
	}))
	defer mockServer.Close()

	var testCases = []struct {
		// What we pass as input
		testArgs Arguments
		// Error matcher (nil if we don't expect an error)
		errorMatcher func(error) bool
	}{
		{
			Arguments{
				APIEndpoint:     mockServer.URL,
				ClusterNameOrID: "v4-cluster-id",
				Workers:         10,
				WorkersSet:      true,
			},
			errors.IsNotLoggedInError,
		},
		{
			Arguments{
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "v4-cluster-id",
				Workers:         10,
				WorkersSet:      true,
			},
			nil,
		},
		{
			Arguments{
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "v4-cluster-id",
				Workers:         10,
				WorkersSet:      true,
				WorkersMin:      4,
				WorkersMinSet:   true,
			},
			errors.IsConflictingWorkerFlagsUsed,
		},
		{
			Arguments{
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "v4-cluster-id",
			},
			errors.IsRequiredFlagMissingError,
		},
		{
			Arguments{
				APIEndpoint:     mockServer.URL,
				AuthToken:       "some-token",
				ClusterNameOrID: "v5-cluster-id",
				Verbose:         true,
				Workers:         10,
				WorkersSet:      true,
			},
			errors.IsCannotScaleCluster,
		},
	}

	var thisConfigYAML = `last_version_check: 0001-01-01T00:00:00Z
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL + `
updated: 2017-09-29T11:23:15+02:00
`

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, thisConfigYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			initFlags()

			clientWrapper, err := client.NewWithConfig(tc.testArgs.APIEndpoint, tc.testArgs.UserProvidedToken)
			if err != nil {
				t.Errorf("Case %d - Unexpected error '%s'", i, err)
			}

			err = verifyPreconditions(tc.testArgs, clientWrapper)
			if tc.errorMatcher == nil {
				if err != nil {
					t.Errorf("Case %d - Unexpected error '%s'", i, err)
				}
			} else {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Expected error %v, got %v", i, runtime.FuncForPC(reflect.ValueOf(tc.errorMatcher).Pointer()).Name(), err)
					if err != nil {
						t.Logf("Case %d - Stack: %s", i, microerror.JSON(err))
					}
				}
			}
		})
	}
}

// TestScaleClusterNotLoggedIn tests if we can prevent an attempt to do things
// when not logged in and no token has been provided.
func TestScaleClusterNotLoggedIn(t *testing.T) {
	// This server should not get any request, because we avoid unauthenticated requests.
	// That's why it issues an error in case it does.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("TestScaleClusterNotLoggedIn mockServer request:", r.Method, r.URL)
		t.Error("A request has been sent although we don't have a token.")
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	configDir := testutils.TempDir(fs)
	config.Initialize(fs, configDir)

	testArgs := Arguments{
		APIEndpoint:     mockServer.URL,
		ClusterNameOrID: "cluster-id",
		Workers:         5,
	}

	clientWrapper, err := client.NewWithConfig(testArgs.APIEndpoint, testArgs.UserProvidedToken)
	if err != nil {
		t.Errorf("Unexpected error '%s'", err)
	}

	err = verifyPreconditions(testArgs, clientWrapper)
	if !errors.IsNotLoggedInError(err) {
		t.Error("Expected NotLoggedInError, got", err)
	}

}

// TestScaleCluster tests scaling a cluster under normal conditions:
// user logged in.
func TestScaleCluster(t *testing.T) {
	testCases := []struct {
		numWorkersBefore int

		initialMinScaling int
		desiredMinScaling int
		workersMinSet     bool
		resultWorkersMin  int

		initialMaxScaling int
		desiredMaxScaling int
		workersMaxSet     bool
		resultWorkersMax  int

		numWorkersAfter int
	}{
		{
			numWorkersBefore: 3,

			initialMinScaling: 3,
			desiredMinScaling: 5,
			workersMinSet:     true,
			resultWorkersMin:  5,

			initialMaxScaling: 3,
			desiredMaxScaling: 5,
			workersMaxSet:     true,
			resultWorkersMax:  5,

			numWorkersAfter: 5,
		},
		{
			numWorkersBefore: 4,

			initialMinScaling: 3,
			desiredMinScaling: 5,
			workersMinSet:     true,
			resultWorkersMin:  5,

			initialMaxScaling: 6,
			desiredMaxScaling: 10,
			workersMaxSet:     true,
			resultWorkersMax:  10,

			numWorkersAfter: 4,
		},
		{
			numWorkersBefore: 4,

			initialMinScaling: 3,
			desiredMinScaling: 5,
			workersMinSet:     true,
			resultWorkersMin:  5,

			initialMaxScaling: 6,
			desiredMaxScaling: 10,
			workersMaxSet:     false,
			resultWorkersMax:  6,

			numWorkersAfter: 4,
		},
		{
			numWorkersBefore: 10,

			initialMinScaling: 8,
			desiredMinScaling: 10,
			workersMinSet:     false,
			resultWorkersMin:  8,

			initialMaxScaling: 8,
			desiredMaxScaling: 10,
			workersMaxSet:     true,
			resultWorkersMax:  10,

			numWorkersAfter: 8,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Logf("mockServer: %s %s", r.Method, r.URL.String())
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "GET" && r.URL.String() == "/v4/clusters/cluster-id/" {
					// cluster details before the patch
					w.WriteHeader(http.StatusOK)
					w.Write(generateClusterResponse(tc.numWorkersBefore, tc.initialMinScaling, tc.initialMaxScaling))
				} else if r.Method == "PATCH" && r.URL.String() == "/v4/clusters/cluster-id/" {
					// inspect PATCH request body
					w.WriteHeader(http.StatusOK)
					patchBytes, readErr := ioutil.ReadAll(r.Body)
					if readErr != nil {
						t.Error(readErr)
					}
					patch, parseErr := gabs.ParseJSON(patchBytes)
					if parseErr != nil {
						t.Error(parseErr)
					}
					if !patch.Exists("workers") {
						t.Error("Patch request body does not contain 'workers' key.")
					}

					w.Write(generateClusterResponse(tc.numWorkersAfter, 0, 0))
				} else if r.Method == "GET" && r.URL.String() == "/v4/clusters/cluster-id/status/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
				"cluster": {
					"conditions": [
						{
							"status": "True",
							"type": "Created"
						}
					],
					"network": {
						"cidr": ""
					},
					"nodes": [
						{
							"name": "4jr2w-master-000000",
							"version": "2.0.1"
						},
						{
							"name": "4jr2w-worker-000001",
							"version": "2.0.1"
						}
					],
					"resources": [],
					"scaling":{
						"desiredCapacity": 3
					},					
					"versions": [
						{
							"date": "0001-01-01T00:00:00Z",
							"semver": "2.0.1"
						}
					]
				}
			}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Could not find this."}`))
				}
			}))
			defer mockServer.Close()

			// config
			yamlText := `
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
    provider: aws
selected_endpoint: ` + mockServer.URL
			fs := afero.NewMemMapFs()
			_, err := testutils.TempConfig(fs, yamlText)
			if err != nil {
				t.Error(err)
			}

			testArgs := Arguments{
				APIEndpoint:       mockServer.URL,
				ClusterNameOrID:   "cluster-id",
				WorkersMax:        int64(tc.desiredMaxScaling),
				WorkersMin:        int64(tc.desiredMinScaling),
				UserProvidedToken: "my-token",
				WorkersMinSet:     tc.workersMinSet,
				WorkersMaxSet:     tc.workersMaxSet,
			}

			clientWrapper, err := client.NewWithConfig(testArgs.APIEndpoint, testArgs.UserProvidedToken)
			if err != nil {
				t.Errorf("Unexpected error '%s'", err)
			}

			err = verifyPreconditions(testArgs, clientWrapper)
			if err != nil {
				t.Error(err)
			}

			result, scaleErr := scaleCluster(testArgs)
			if scaleErr != nil {
				t.Error(scaleErr)
			}

			expectedResult := Result{
				NumWorkersBefore: tc.numWorkersBefore,
				ScalingMinBefore: tc.initialMinScaling,
				ScalingMinAfter:  tc.resultWorkersMin,
				ScalingMaxBefore: tc.initialMaxScaling,
				ScalingMaxAfter:  tc.resultWorkersMax,
			}
			if diff := cmp.Diff(expectedResult, *result); diff != "" {
				t.Errorf("Case %d: Scaling result unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

func generateClusterResponse(workerCount int, workersMin int, workersMax int) []byte {
	c := models.V4ClusterDetailsResponse{
		ID:          "cluster-id",
		Name:        "",
		APIEndpoint: "",
		CreateDate:  "2017-05-16T09:30:31.192170835Z",
		Owner:       "acmeorg",
		Scaling: &models.V4ClusterDetailsResponseScaling{
			Min: int64(workersMin),
			Max: int64(workersMax),
		},
	}

	for i := 0; i < workerCount; i++ {
		c.Workers = append(c.Workers, &models.V4ClusterDetailsResponseWorkersItems{
			Memory: &models.V4ClusterDetailsResponseWorkersItemsMemory{
				SizeGb: 5,
			},
			Storage: &models.V4ClusterDetailsResponseWorkersItemsStorage{
				SizeGb: 5,
			},
			CPU: &models.V4ClusterDetailsResponseWorkersItemsCPU{
				Cores: 2,
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		})
	}

	bytes, _ := json.Marshal(c)

	return bytes
}
