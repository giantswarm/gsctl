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

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
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
			description: "Definition from v4 YAML file",
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
