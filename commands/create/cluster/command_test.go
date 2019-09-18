package cluster

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/testutils"
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

// Test_CollectArgsV4 tests whether collectArguments produces the expected results.
func Test_CollectArgsV4(t *testing.T) {
	var testCases = []struct {
		// The flags we pass to the command.
		flags []string
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{""},
			Arguments{
				APIEndpoint: "https://foo",
				AuthToken:   "some-token",
				Scheme:      "giantswarm",
			},
		},
		{
			[]string{
				"--owner=acme",
				"--name=ClusterName",
				"--release=1.2.3",
			},
			Arguments{
				APIEndpoint:    "https://foo",
				AuthToken:      "some-token",
				ClusterName:    "ClusterName",
				Owner:          "acme",
				ReleaseVersion: "1.2.3",
				Scheme:         "giantswarm",
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
			Command.ParseFlags(tc.flags)

			args := collectArguments()
			if err != nil {
				t.Errorf("Case %d - Unexpected error '%s'", i, err)
			}
			if diff := cmp.Diff(tc.resultingArgs, args, cmpopts.IgnoreFields(Arguments{}, "FileSystem")); diff != "" {
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
		// Token missing.
		{
			Arguments{
				APIEndpoint: "https://mock-url",
			},
			errors.IsNotLoggedInError,
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

// Test_ReadDefinitionFiles tests the readDefinitionFromFile with all
// YAML files in the testdata directory.
func Test_ReadDefinitionFiles(t *testing.T) {
	basePath := "testdata"
	fs := afero.NewOsFs()
	files, _ := afero.ReadDir(fs, basePath)

	for i, f := range files {
		t.Logf("Case %d, file %s", i, f.Name())
		path := basePath + "/" + f.Name()
		_, err := readDefinitionFromFile(fs, path)
		if err != nil {
			t.Errorf("Unexpected error in case %d, file %s: %s", i, f.Name(), err)
		}
	}
}

// Test_ParseYAMLDefinitionV4 tests parsing v4 YAML definition files.
func Test_ParseYAMLDefinitionV4(t *testing.T) {
	var testCases = []struct {
		inputYAML      []byte
		expectedOutput *types.ClusterDefinitionV4
	}{
		// Minimal YAML.
		{
			[]byte(`owner: myorg`),
			&types.ClusterDefinitionV4{
				Owner: "myorg",
			},
		},
		// More details.
		{
			[]byte(`owner: myorg
name: My cluster
release_version: 1.2.3
availability_zones: 3
scaling:
  min: 3
  max: 5`),
			&types.ClusterDefinitionV4{
				Owner:             "myorg",
				Name:              "My cluster",
				ReleaseVersion:    "1.2.3",
				AvailabilityZones: 3,
				Scaling: types.ScalingDefinition{
					Min: 3,
					Max: 5,
				},
			},
		},
		// KVM worker details.
		{
			[]byte(`owner: myorg
workers:
- memory:
    size_gb: 16.5
  cpu:
    cores: 4
  storage:
    size_gb: 100
- memory:
    size_gb: 32
  cpu:
    cores: 8
  storage:
    size_gb: 50
`),
			&types.ClusterDefinitionV4{
				Owner: "myorg",
				Workers: []types.NodeDefinition{
					types.NodeDefinition{
						Memory:  types.MemoryDefinition{SizeGB: 16.5},
						CPU:     types.CPUDefinition{Cores: 4},
						Storage: types.StorageDefinition{SizeGB: 100},
					},
					types.NodeDefinition{
						Memory:  types.MemoryDefinition{SizeGB: 32},
						CPU:     types.CPUDefinition{Cores: 8},
						Storage: types.StorageDefinition{SizeGB: 50},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			def, err := readDefinitionFromYAML(tc.inputYAML)
			if err != nil {
				t.Errorf("Case %d - Unexpected error %v", i, err)
			}

			if diff := cmp.Diff(tc.expectedOutput, def); diff != "" {
				t.Errorf("Case %d - Resulting definition unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// Test_ParseYAMLDefinitionV5 tests parsing v5 YAML definition files.
func Test_ParseYAMLDefinitionV5(t *testing.T) {
	var testCases = []struct {
		inputYAML      []byte
		expectedOutput *types.ClusterDefinitionV5
	}{
		// Minimal YAML.
		{
			[]byte(`api_version: v5
owner: myorg`),
			&types.ClusterDefinitionV5{
				Owner: "myorg",
			},
		},
		// More details.
		{
			[]byte(`api_version: v5
owner: myorg
name: My cluster
release_version: 1.2.3
`),
			&types.ClusterDefinitionV5{
				Owner:          "myorg",
				Name:           "My cluster",
				ReleaseVersion: "1.2.3",
			},
		},
		// Node pools.
		{
			[]byte(`api_version: v5
owner: myorg
master:
  availability_zone: my-zone-1a
nodepools:
- name: General purpose
  availability_zones:
    number: 2
- name: Database
  availability_zones:
    zones:
    - my-zone-1a
    - my-zone-1b
    - my-zone-1c
  scaling:
    min: 3
    max: 10
  node_spec:
    aws:
      instance_type: "m5.superlarge"
- name: Batch
`),
			&types.ClusterDefinitionV5{
				Owner:  "myorg",
				Master: &types.MasterDefinition{AvailabilityZone: "my-zone-1a"},
				NodePools: []*types.NodePoolDefinition{
					&types.NodePoolDefinition{
						Name:              "General purpose",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Number: 2},
					},
					&types.NodePoolDefinition{
						Name:              "Database",
						AvailabilityZones: &types.AvailabilityZonesDefinition{Zones: []string{"my-zone-1a", "my-zone-1b", "my-zone-1c"}},
						Scaling:           &types.ScalingDefinition{Min: 3, Max: 10},
						NodeSpec:          &types.NodeSpec{AWS: &types.AWSSpecificDefinition{InstanceType: "m5.superlarge"}},
					},
					&types.NodePoolDefinition{
						Name: "Batch",
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			def, err := readDefinitionFromYAML(tc.inputYAML)
			if err != nil {
				t.Errorf("Case %d - Unexpected error %v", i, err)
			}

			if diff := cmp.Diff(tc.expectedOutput, def); diff != "" {
				t.Errorf("Case %d - Resulting definition unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}

// Test_CreateFromBadYAML01 tests how non-conforming YAML is treated.
func Test_CreateFromBadYAML01(t *testing.T) {
	data := []byte(`o: myorg`)
	def := types.ClusterDefinitionV4{}

	err := yaml.Unmarshal(data, &def)
	if err != nil {
		t.Fatalf("expected error to be empty, got %#v", err)
	}

	if def.Owner != "" {
		t.Fatalf("expected owner to be empty, got %q", def.Owner)
	}
}

// Test_CreateClusterSuccessfully tests cluster creations that should succeed.
func Test_CreateClusterSuccessfully(t *testing.T) {
	var testCases = []struct {
		description string
		inputArgs   *Arguments
	}{
		{
			description: "Minimal arguments",
			inputArgs: &Arguments{
				Owner:     "acme",
				AuthToken: "fake token",
			},
		},
		{
			description: "Extensive arguments",
			inputArgs: &Arguments{
				ClusterName:    "UnitTestCluster",
				ReleaseVersion: "0.3.0",
				Owner:          "acme",
				AuthToken:      "fake token",
				Verbose:        true,
			},
		},
		{
			description: "Definition from YAML file",
			inputArgs: &Arguments{
				ClusterName:   "Cluster Name from Args",
				FileSystem:    afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile: "testdata/minimal.yaml",
				Owner:         "acme",
				AuthToken:     "fake token",
				Verbose:       true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Logf("Case %d: %s", i, testCase.description)

		// mock server always responding positively
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("mockServer request: ", r.Method, r.URL)
			w.Header().Set("Content-Type", "application/json")
			if !strings.Contains(r.Header.Get("Authorization"), testCase.inputArgs.AuthToken) {
				t.Errorf("Authorization header incomplete: '%s'", r.Header.Get("Authorization"))
			}
			if r.Method == "POST" && r.URL.String() == "/v4/clusters/" {
				w.Header().Set("Location", "/v4/clusters/f6e8r/")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"code": "RESOURCE_CREATED", "message": "Yeah!"}`))
			} else if r.Method == "GET" && r.URL.String() == "/v4/info/" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"general": {
					  "provider": "aws"
					},
					"features": {
					  "nodepools": {"release_version_minimum": "9.0.0"}
					}
				  }`))
			} else if r.Method == "GET" && r.URL.String() == "/v4/releases/" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
			  {
					"timestamp": "2017-10-15T12:00:00Z",
			    "version": "0.3.0",
			    "active": true,
			    "changelog": [
			      {
			        "component": "firstComponent",
			        "description": "firstComponent added."
			      }
			    ],
			    "components": [
			      {
			        "name": "firstComponent",
			        "version": "0.0.1"
			      }
			    ]
			  }
			]`))
			}
		}))
		defer mockServer.Close()

		testCase.inputArgs.APIEndpoint = mockServer.URL
		testCase.inputArgs.UserProvidedToken = testCase.inputArgs.AuthToken

		err := verifyPreconditions(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Validation error in testCase %d: %s", i, err.Error())
		}
		_, err = addCluster(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Execution error in testCase %d: %s", i, err.Error())
		}
	}
}

// Test_CreateClusterExecutionFailures tests for errors thrown in the
// final execution of a cluster creations, which is the handling of the API call.
func Test_CreateClusterExecutionFailures(t *testing.T) {
	var testCases = []struct {
		description        string
		inputArgs          *Arguments
		responseStatus     int
		serverResponseJSON []byte
		errorMatcher       func(err error) bool
	}{
		{
			description: "Unauthenticated request despite token being present",
			inputArgs: &Arguments{
				Owner:     "owner",
				AuthToken: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "PERMISSION_DENIED", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusUnauthorized,
			errorMatcher:       errors.IsNotAuthorizedError,
		},
		{
			description: "Owner organization not existing",
			inputArgs: &Arguments{
				Owner:     "non-existing-owner",
				AuthToken: "some-token",
			},
			serverResponseJSON: []byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Lorem ipsum"}`),
			responseStatus:     http.StatusNotFound,
			errorMatcher:       errors.IsOrganizationNotFoundError,
		},
		{
			description: "Non-existing YAML definition path",
			inputArgs: &Arguments{
				Owner:         "owner",
				AuthToken:     "some-token",
				FileSystem:    afero.NewOsFs(),
				InputYAMLFile: "does/not/exist.yaml",
			},
			serverResponseJSON: []byte(``),
			responseStatus:     400,
			errorMatcher:       errors.IsYAMLFileNotReadable,
		},
	}

	for i, testCase := range testCases {
		t.Logf("Case %d: %s", i, testCase.description)

		// mock server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//t.Log("mockServer request: ", r.Method, r.URL)
			if r.Method == "GET" && r.URL.String() == "/v4/info/" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"general": {
					  "provider": "aws"
					},
					"features": {
					  "nodepools": {"release_version_minimum": "9.0.0"}
					}
				  }`))
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(testCase.responseStatus)
				w.Write([]byte(testCase.serverResponseJSON))
			}
		}))
		defer mockServer.Close()

		// client
		flags.APIEndpoint = mockServer.URL // required to make InitClient() work
		testCase.inputArgs.APIEndpoint = mockServer.URL

		err := verifyPreconditions(*testCase.inputArgs)
		if err != nil {
			t.Errorf("Unexpected error in argument validation: %#v", err)
		} else {
			_, err := addCluster(*testCase.inputArgs)
			if err == nil {
				t.Errorf("Test case %d did not yield an execution error.", i)
			}
			origErr := microerror.Cause(err)
			if testCase.errorMatcher(origErr) == false {
				t.Errorf("Test case %d did not yield the expected execution error, instead: %#v", i, err)
			}
		}
	}
}

func Test_getLatestActiveReleaseVersion(t *testing.T) {
	var testCases = []struct {
		responseBody  string
		latestRelease string
		errorMatcher  func(err error) bool
	}{
		{
			responseBody: `[
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "0.3.0",
					"active": true,
					"changelog": [],
					"components": []
				},
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "0.2.1",
					"active": true,
					"changelog": [],
					"components": []
				}
			]`,
			latestRelease: "0.3.0",
		},
		{
			responseBody: `[
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "0.3.0",
					"active": true,
					"changelog": [],
					"components": []
				},
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "1.6.1",
					"active": true,
					"changelog": [],
					"components": []
				},
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "2.3.0",
					"active": false,
					"changelog": [],
					"components": []
				}
			  ]`,
			latestRelease: "1.6.1",
		},
	}

	for i, tc := range testCases {

		// mock server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//t.Log("mockServer request: ", r.Method, r.URL)
			if r.Method == "GET" && r.URL.String() == "/v4/releases/" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tc.responseBody))
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code": "UNKNOWN_ERROR", "message": "Can't do this"}`))
			}
		}))
		defer mockServer.Close()

		// client
		flags.APIEndpoint = mockServer.URL // required to make InitClient() work

		args := &Arguments{
			APIEndpoint: mockServer.URL,
			AuthToken:   "some-token",
		}

		clientWrapper, err := client.NewWithConfig(args.APIEndpoint, args.UserProvidedToken)
		if err != nil {
			t.Errorf("Test case %d: Error %s", i, err)
		}

		latest, err := getLatestActiveReleaseVersion(clientWrapper, nil)
		if err != nil {
			t.Errorf("Test case %d: Error %s", i, err)
		}

		if latest != tc.latestRelease {
			t.Errorf("Test case %d: Expected '%s' but got '%s'", i, tc.latestRelease, latest)
		}
	}
}
