package cluster

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"

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

// Test_CollectArgs tests whether collectArguments produces the expected results.
func Test_CollectArgs(t *testing.T) {
	var testCases = []struct {
		// The flags we pass to the command.
		flags []string
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{""},
			Arguments{
				APIEndpoint:           "https://foo",
				AuthToken:             "some-token",
				CreateDefaultNodePool: true,
				Scheme:                "giantswarm",
				MasterHA:              nil,
			},
		},
		{
			[]string{"--master-ha=false"},
			Arguments{
				APIEndpoint:           "https://foo",
				AuthToken:             "some-token",
				CreateDefaultNodePool: true,
				Scheme:                "giantswarm",
				MasterHA:              toBoolPtr(false),
			},
		},
		{
			[]string{
				"--owner=acme",
				"--name=ClusterName",
				"--release=1.2.3",
			},
			Arguments{
				APIEndpoint:           "https://foo",
				AuthToken:             "some-token",
				ClusterName:           "ClusterName",
				CreateDefaultNodePool: true,
				Owner:                 "acme",
				ReleaseVersion:        "1.2.3",
				Scheme:                "giantswarm",
				MasterHA:              nil,
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

			args := collectArguments(Command)
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

func stringP(v string) *string {
	return &v
}

// Test_CreateClusterSuccessfully tests cluster creations that should succeed.
func Test_CreateClusterSuccessfully(t *testing.T) {
	var testCases = []struct {
		description    string
		inputArgs      *Arguments
		expectedResult *creationResult
	}{
		{
			description: "Minimal arguments",
			inputArgs: &Arguments{
				Owner:     "acme",
				AuthToken: "fake token",
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					Owner: "acme",
					MasterNodes: &types.MasterNodes{
						HighAvailability: true,
					},
				},
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
			expectedResult: &creationResult{
				ID:       "f6e8r",
				Location: "/v4/clusters/f6e8r/",
				DefinitionV4: &types.ClusterDefinitionV4{
					Name:           "UnitTestCluster",
					Owner:          "acme",
					ReleaseVersion: "0.3.0",
				},
			},
		},
		{
			description: "Definition from minimal v4 YAML file, release version via args",
			inputArgs: &Arguments{
				ClusterName:    "Cluster Name from Args",
				FileSystem:     afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile:  "testdata/v4_minimal.yaml",
				Owner:          "acme",
				AuthToken:      "fake token",
				ReleaseVersion: "1.0.0",
				Verbose:        true,
			},
			expectedResult: &creationResult{
				ID:       "f6e8r",
				Location: "/v4/clusters/f6e8r/",
				DefinitionV4: &types.ClusterDefinitionV4{
					Name:           "Cluster Name from Args",
					Owner:          "acme",
					ReleaseVersion: "1.0.0",
				},
			},
		},
		{
			description: "Definition from complex v4 YAML file",
			inputArgs: &Arguments{
				FileSystem:    afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile: "testdata/v4_complete.yaml",
				Owner:         "acme",
				AuthToken:     "fake token",
				Verbose:       true,
			},
			expectedResult: &creationResult{
				ID:       "f6e8r",
				Location: "/v4/clusters/f6e8r/",
				DefinitionV4: &types.ClusterDefinitionV4{
					Name:              "Complete cluster spec",
					Owner:             "acme",
					ReleaseVersion:    "1.2.3",
					AvailabilityZones: 1,
					Workers: []types.NodeDefinition{
						{
							Memory:  types.MemoryDefinition{SizeGB: 2},
							CPU:     types.CPUDefinition{Cores: 2},
							Storage: types.StorageDefinition{SizeGB: 20},
							Labels:  map[string]string{"nodetype": "standard"},
						},
						{
							Memory:  types.MemoryDefinition{SizeGB: 8},
							CPU:     types.CPUDefinition{Cores: 2},
							Storage: types.StorageDefinition{SizeGB: 20},
							Labels:  map[string]string{"nodetype": "hiram"},
						},
						{
							Memory:  types.MemoryDefinition{SizeGB: 2},
							CPU:     types.CPUDefinition{Cores: 6},
							Storage: types.StorageDefinition{SizeGB: 20},
							Labels:  map[string]string{"nodetype": "hicpu"},
						},
					},
				},
			},
		},
		{
			description: "Definition from minimal v5 YAML file, no release version",
			inputArgs: &Arguments{
				ClusterName:           "Cluster Name from Args",
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile:         "testdata/v5_minimal.yaml",
				Owner:                 "acme",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion: "v5",
					Name:       "Cluster Name from Args",
					Owner:      "acme",
					MasterNodes: &types.MasterNodes{
						HighAvailability: true,
					},
				},
			},
		},
		{
			description: "Definition from complex v5 YAML file",
			inputArgs: &Arguments{
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(),
				InputYAMLFile:         "testdata/v5_three_nodepools.yaml",
				Owner:                 "acme",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion:     "v5",
					Name:           "Cluster with three node pools",
					Owner:          "acme",
					ReleaseVersion: "9.0.0",
					Master:         &types.MasterDefinition{AvailabilityZone: "eu-central-1a"},
					NodePools: []*types.NodePoolDefinition{
						{
							Name:              "Node pool with 2 random AZs",
							AvailabilityZones: &types.AvailabilityZonesDefinition{Number: 2},
						},
						{
							Name: "Node pool with 3 specific AZs A, B, C, scaling 3-10, m5.xlarge",
							AvailabilityZones: &types.AvailabilityZonesDefinition{
								Zones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
							},
							Scaling: &types.ScalingDefinition{
								Min: 3,
								Max: 10,
							},
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceType: "m5.xlarge",
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                0,
										OnDemandPercentageAboveBaseCapacity: 100,
									},
									UseAlikeInstanceTypes: true,
								},
							},
						},
						{
							Name: "Node pool using defaults only",
						},
					},
				},
			},
		},
		{
			description: "Definition with several spot-related node pool setups (v5_instance_distribution.yaml)",
			inputArgs: &Arguments{
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(),
				InputYAMLFile:         "testdata/v5_instance_distribution.yaml",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion:     "v5",
					Name:           "Cluster with several node pools testing various instance distribution combinations",
					Owner:          "acme",
					ReleaseVersion: "11.5.0",
					MasterNodes: &types.MasterNodes{
						HighAvailability: true,
					},
					NodePools: []*types.NodePoolDefinition{
						{
							Name: "Node pool with 0 on-demand, 100% spot, no alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                0,
										OnDemandPercentageAboveBaseCapacity: 100,
									},
									UseAlikeInstanceTypes: false,
								},
							},
						},
						{
							Name: "Node pool with 3 on-demand, 100% spot, no alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                3,
										OnDemandPercentageAboveBaseCapacity: 100,
									},
									UseAlikeInstanceTypes: false,
								},
							},
						},
						{
							Name: "Node pool with 3 on-demand, 50% spot, no alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                3,
										OnDemandPercentageAboveBaseCapacity: 50,
									},
									UseAlikeInstanceTypes: false,
								},
							},
						},
						{
							Name: "Node pool with 0 on-demand, 100% spot, use alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                0,
										OnDemandPercentageAboveBaseCapacity: 100,
									},
									UseAlikeInstanceTypes: true,
								},
							},
						},
						{
							Name: "Node pool with 3 on-demand, 100% spot, use alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                3,
										OnDemandPercentageAboveBaseCapacity: 100,
									},
									UseAlikeInstanceTypes: true,
								},
							},
						},
						{
							Name: "Node pool with 3 on-demand, 50% spot, use alike instance types",
							NodeSpec: &types.NodeSpec{
								AWS: &types.AWSSpecificDefinition{
									InstanceDistribution: &types.AWSInstanceDistribution{
										OnDemandBaseCapacity:                3,
										OnDemandPercentageAboveBaseCapacity: 50,
									},
									UseAlikeInstanceTypes: true,
								},
							},
						},
					},
				},
			},
		},
		{
			description: "Definition from v5 YAML file with labels",
			inputArgs: &Arguments{
				ClusterName:           "Cluster Name from Args with Labels",
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile:         "testdata/v5_with_labels.yaml",
				Owner:                 "acme",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion: "v5",
					Name:       "Cluster Name from Args with Labels",
					Owner:      "acme",
					MasterNodes: &types.MasterNodes{
						HighAvailability: true,
					},
					Labels: map[string]*string{"key": stringP("value"), "labelkey": stringP("labelvalue")},
				},
			},
		},
		{
			description: "Definition from v5 YAML file with HA master",
			inputArgs: &Arguments{
				ClusterName:           "Cluster Name from Args with Labels",
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile:         "testdata/v5_with_ha_master.yaml",
				Owner:                 "acme",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion: "v5",
					Name:       "Cluster Name from Args with Labels",
					Owner:      "acme",
					MasterNodes: &types.MasterNodes{
						HighAvailability: true,
					},
				},
			},
		},
		{
			description: "Definition from v5 YAML file with HA master, overridden by flag",
			inputArgs: &Arguments{
				ClusterName:           "Cluster Name from Args with Labels",
				CreateDefaultNodePool: false,
				FileSystem:            afero.NewOsFs(), // needed for YAML file access
				InputYAMLFile:         "testdata/v5_with_ha_master.yaml",
				MasterHA:              toBoolPtr(false),
				Owner:                 "acme",
				AuthToken:             "fake token",
				Verbose:               true,
			},
			expectedResult: &creationResult{
				ID: "f6e8r",
				DefinitionV5: &types.ClusterDefinitionV5{
					APIVersion: "v5",
					Name:       "Cluster Name from Args with Labels",
					Owner:      "acme",
					MasterNodes: &types.MasterNodes{
						HighAvailability: false,
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("Case %d: %s", i, tc.description)

			fs := afero.NewMemMapFs()
			_, err := testutils.TempConfig(fs, configYAML)
			if err != nil {
				t.Fatal(err)
			}

			// mock server always responding positively
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Log("mockServer request: ", r.Method, r.URL)
				w.Header().Set("Content-Type", "application/json")
				if !strings.Contains(r.Header.Get("Authorization"), tc.inputArgs.AuthToken) {
					t.Errorf("Authorization header incomplete: '%s'", r.Header.Get("Authorization"))
				}
				if r.Method == "POST" && r.URL.String() == "/v4/clusters/" {
					w.Header().Set("Location", "/v4/clusters/f6e8r/")
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"code": "RESOURCE_CREATED", "message": "Yeah!"}`))
				} else if r.Method == "POST" && r.URL.String() == "/v5/clusters/" {
					w.Header().Set("Location", "/v5/clusters/f6e8r/")
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{
					"id": "f6e8r",
					"owner": "acme",
					"release_version": "9.0.0",
					"name": "Node Pool Cluster",
					"master": {
						"availability_zone": "eu-central-1c"
					},
					"nodepools": [],
					"labels": {
						"key": "value",
						"labelkey": "labelvalue"
					}
				}`))
				} else if r.Method == "GET" && r.URL.String() == "/v4/info/" {
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
				} else if r.Method == "GET" && r.URL.String() == "/v4/releases/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
					{
						"timestamp": "2019-01-01T12:00:00Z",
						"version": "1.0.0",
						"active": true,
						"changelog": [],
						"components": []
					},
					{
						"timestamp": "2019-09-23T12:00:00Z",
						"version": "9.0.0",
						"active": true,
						"changelog": [],
						"components": []
					},
					{
						"timestamp": "2019-10-23T12:00:00Z",
						"version": "11.5.0",
						"active": true,
						"changelog": [],
						"components": []
					}
				]`))
				} else if r.Method == "POST" && r.URL.String() == "/v5/clusters/f6e8r/nodepools/" {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{
					"id": "a1b2",
					"name": "Default node pool name",
					"availability_zones": ["eu-central-1a"],
					"scaling": {"min": 3, "max": 3},
					"node_spec": {
					  "aws": {
						"instance_type": "m4.2xlarge"
					  }
					}
				  }`))
				}
			}))
			defer mockServer.Close()

			tc.inputArgs.APIEndpoint = mockServer.URL
			tc.inputArgs.UserProvidedToken = tc.inputArgs.AuthToken

			err = verifyPreconditions(*tc.inputArgs)
			if err != nil {
				t.Errorf("Case %d - Validation error: %s", i, err.Error())
			}

			result, err := addCluster(*tc.inputArgs)
			if err != nil {
				t.Errorf("Case %d - Execution error: %s", i, err.Error())
			}

			if diff := cmp.Diff(tc.expectedResult, result, nil); diff != "" {
				t.Errorf("Case %d - Results unequal (-expected +got):\n%s", i, diff)
			}
		})
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

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Fatal(err)
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("Case %d: %s", i, testCase.description)

			// mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// t.Log("mockServer request: ", r.Method, r.URL)
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
					w.Write(testCase.serverResponseJSON)
				}
			}))
			defer mockServer.Close()

			// client
			flags.APIEndpoint = mockServer.URL // required to make InitClient() work
			testCase.inputArgs.APIEndpoint = mockServer.URL
			testCase.inputArgs.FileSystem = fs

			err := verifyPreconditions(*testCase.inputArgs)
			if err != nil {
				t.Errorf("Unexpected error in argument validation: %#v", err)
			} else {
				_, err := addCluster(*testCase.inputArgs)
				if err == nil {
					t.Errorf("Test case %d did not yield an execution error.", i)
				}
				if testCase.errorMatcher(err) == false {
					t.Errorf("Test case %d did not yield the expected execution error, instead: %#v", i, err)
				}
			}
		})
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
		{
			responseBody: `[
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "",
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
					"version": "quirks",
					"active": true,
					"changelog": [],
					"components": []
				},
				{
					"timestamp": "2017-10-15T12:00:00Z",
					"version": "2.3.0",
					"active": true,
					"changelog": [],
					"components": []
				}
			  ]`,
			latestRelease: "2.3.0",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// t.Log("mockServer request: ", r.Method, r.URL)
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
		})
	}
}

func Test_validateHAMasters(t *testing.T) {
	testCases := []struct {
		name           string
		featureEnabled bool
		args           Arguments
		v5Definition   types.ClusterDefinitionV5

		errorMatcher       func(error) bool
		expectedResultArgs Arguments
	}{
		{
			name:           "Use default arguments, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master:      nil,
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Use default arguments, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master:      nil,
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: toBoolPtr(true),
			},
		},
		{
			name:           "Set specific availability zone, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set specific availability zone, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set HA master nodes, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: nil,
				MasterNodes: &types.MasterNodes{
					HighAvailability: true,
				},
			},
			errorMatcher: IsHAMastersNotSupported,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set HA master nodes, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: nil,
				MasterNodes: &types.MasterNodes{
					HighAvailability: true,
				},
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set master AZ, and HA master, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: &types.MasterNodes{
					HighAvailability: true,
				},
			},
			errorMatcher: IsMustProvideSingleMasterType,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set master AZ, and HA master, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: &types.MasterNodes{
					HighAvailability: true,
				},
			},
			errorMatcher: IsMustProvideSingleMasterType,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set master AZ, and HA master through flag, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: toBoolPtr(true),
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: nil,
			},
			errorMatcher: IsMustProvideSingleMasterType,
			expectedResultArgs: Arguments{
				MasterHA: toBoolPtr(true),
			},
		},
		{
			name:           "Set master AZ, and HA master through flag, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: toBoolPtr(true),
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: &types.MasterDefinition{
					AvailabilityZone: "eu-something",
				},
				MasterNodes: nil,
			},
			errorMatcher: IsMustProvideSingleMasterType,
			expectedResultArgs: Arguments{
				MasterHA: toBoolPtr(true),
			},
		},
		{
			name:           "Set HA master to false, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: nil,
				MasterNodes: &types.MasterNodes{
					HighAvailability: false,
				},
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set HA master to false, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: nil,
			},
			v5Definition: types.ClusterDefinitionV5{
				Master: nil,
				MasterNodes: &types.MasterNodes{
					HighAvailability: false,
				},
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: nil,
			},
		},
		{
			name:           "Set HA master to false through flag, HA masters turned off",
			featureEnabled: false,
			args: Arguments{
				MasterHA: toBoolPtr(false),
			},
			v5Definition: types.ClusterDefinitionV5{
				Master:      nil,
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: toBoolPtr(false),
			},
		},
		{
			name:           "Set HA master to false through flag, HA masters turned on",
			featureEnabled: true,
			args: Arguments{
				MasterHA: toBoolPtr(false),
			},
			v5Definition: types.ClusterDefinitionV5{
				Master:      nil,
				MasterNodes: nil,
			},
			errorMatcher: nil,
			expectedResultArgs: Arguments{
				MasterHA: toBoolPtr(false),
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			err := validateHAMasters(tc.featureEnabled, &tc.args, &tc.v5Definition)

			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Unexpected error: %s", i, err)
				}
			} else if err != nil {
				t.Errorf("Case %d - Unexpected error: %s", i, err)
			}

			if diff := cmp.Diff(tc.expectedResultArgs, tc.args); diff != "" {
				t.Errorf("Case %d - Resulting args unequal. (-expected +got):\n%s", i, diff)
			}
		})
	}
}
