package nodepool

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/pkg/provider"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
)

// TestCollectArgs tests whether collectArguments produces the expected results.
func TestCollectArgs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch uri := r.URL.Path; uri {
			case "/v4/info/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"general": {
						"installation_name": "codename",
						"provider": "aws",
						"datacenter": "myzone",
						"availability_zones": {
						  "default": 3,
						  "max": 3
					  	}
					},
					"workers": {
						"count_per_cluster": {
							"max": 20,
							"default": 3
						},
						"instance_type": {
							"options": ["m3.large", "m4.xlarge"],
							"default": "m3.large"
						}
					}
				}`))
			default:
				t.Errorf("Unsupported route %s called in mock server", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
			}
		} else {
			t.Errorf("Unsupported method %s called in mock server", r.Method)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
		}
	}))
	defer mockServer.Close()

	var testCases = []struct {
		// The positional arguments we pass.
		positionalArguments []string
		// How we execute the command.
		commandExecution func()
		// What we expect as arguments.
		resultingArgs Arguments
	}{
		{
			[]string{"cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{"cluster-id", "--name=my-name"})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "cluster-id",
				Name:                       "my-name",
				Scheme:                     "giantswarm",
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
		{
			[]string{"some-cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{
					"some-cluster-id",
					"--name=my-nodepool-name",
					"--num-availability-zones=3",
					"--aws-instance-type=instance-type",
					"--nodes-min=5",
					"--nodes-max=10",
				})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "some-cluster-id",
				Name:                       "my-nodepool-name",
				Scheme:                     "giantswarm",
				AvailabilityZonesNum:       3,
				InstanceType:               "instance-type",
				ScalingMin:                 5,
				ScalingMinSet:              true,
				ScalingMax:                 10,
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
		{
			[]string{"a-cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{
					"a-cluster-id",
					"--availability-zones=A,B,c",
				})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "a-cluster-id",
				Scheme:                     "giantswarm",
				AvailabilityZonesList:      []string{"myzonea", "myzoneb", "myzonec"},
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
		// Only setting the --nodes-min, but not --nodes-max flag.
		{
			[]string{"another-cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{
					"another-cluster-id",
					"--nodes-min=5",
				})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "another-cluster-id",
				ScalingMax:                 0,
				ScalingMin:                 5,
				ScalingMinSet:              true,
				Scheme:                     "giantswarm",
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
		// Only setting the --nodes-max, but not --nodes-min flag.
		{
			[]string{"another-cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{
					"another-cluster-id",
					"--nodes-max=5",
				})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "another-cluster-id",
				ScalingMax:                 5,
				Scheme:                     "giantswarm",
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
		// Setting the Azure VM size.
		{
			[]string{"another-cluster-id"},
			func() {
				initFlags()
				Command.ParseFlags([]string{
					"another-cluster-id",
					"--azure-vm-size=something-large",
				})
			},
			Arguments{
				APIEndpoint:                mockServer.URL,
				AuthToken:                  "some-token",
				ClusterNameOrID:            "another-cluster-id",
				VmSize:                     "something-large",
				Scheme:                     "giantswarm",
				Provider:                   "aws",
				MaxNumOfAvailabilityZones:  3,
				AzureSpotInstancesMaxPrice: -1,
			},
		},
	}

	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
	if err != nil {
		t.Fatal(err)
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

// TestSuccess tests node pool creation with cases that are expected to succeed.
func TestSuccess(t *testing.T) {
	var testCases = []struct {
		args         Arguments
		responseBody string
	}{
		// Minimal node pool creation.
		{
			Arguments{
				ClusterNameOrID: "cluster-id",
				AuthToken:       "token",
				Provider:        "aws",
			},
			`{
				"id": "m0ckr",
				"name": "Unnamed node pool 1",
				"availability_zones": ["eu-central-1a"],
				"scaling": {"min": 3, "max": 3},
				"node_spec": {"aws": {"instance_type": "m5.large"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 0, "nodes_ready": 0},
				"subnet": "10.1.0.0/24"
			}`,
		},
		// Creation with availability zones list.
		{
			Arguments{
				ClusterNameOrID:       "cluster-id",
				AuthToken:             "token",
				Name:                  "my node pool",
				ScalingMin:            4,
				ScalingMax:            10,
				InstanceType:          "my-big-type",
				AvailabilityZonesList: []string{"my-region-1a", "my-region-1c"},
				Provider:              "aws",
			},
			`{
				"id": "m0ckr",
				"name": "my node pool",
				"availability_zones": ["my-region-1a", "my-region-1c"],
				"scaling": {"min": 4, "max": 10},
				"node_spec": {"aws": {"instance_type": "my-big-type"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 0, "nodes_ready": 0},
				"subnet": "10.1.0.0/24"
			}`,
		},
		// Creation with availability zones number.
		{
			Arguments{
				ClusterNameOrID:      "cluster-id",
				AuthToken:            "token",
				Name:                 "my node pool",
				ScalingMin:           2,
				ScalingMax:           50,
				InstanceType:         "my-big-type",
				AvailabilityZonesNum: 3,
				Provider:             "aws",
			},
			`{
				"id": "m0ckr",
				"name": "my node pool",
				"availability_zones": ["my-region-1a", "my-region-1b", "my-region-1c"],
				"scaling": {"min": 4, "max": 10},
				"node_spec": {"aws": {"instance_type": "my-big-type"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 0, "nodes_ready": 0},
				"subnet": "10.1.0.0/24"
			}`,
		},
		// Creation with Azure VM Size.
		{
			Arguments{
				ClusterNameOrID:            "cluster-id",
				AuthToken:                  "token",
				Name:                       "my node pool",
				VmSize:                     "my-big-type",
				Provider:                   "azure",
				AzureSpotInstancesMaxPrice: -1,
			},
			`{
				"id": "m0ckr",
				"name": "my node pool",
				"availability_zones": ["my-region-1a", "my-region-1b", "my-region-1c"],
				"node_spec": {"azure": {"vm_size": "my-big-type"}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 0, "nodes_ready": 0},
				"subnet": "10.1.0.0/24"
			}`,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// ste up mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "POST" {
					switch uri := r.URL.Path; uri {
					case "/v5/clusters/cluster-id/nodepools/":
						w.WriteHeader(http.StatusCreated)
						w.Write([]byte(tc.responseBody))
					default:
						t.Errorf("Unsupported route %s called in mock server", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
					}
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			err := verifyPreconditions(tc.args)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			clientWrapper, err := client.NewWithConfig(tc.args.APIEndpoint, tc.args.UserProvidedToken)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			r, err := createNodePool(tc.args, tc.args.ClusterNameOrID, clientWrapper)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			if r.nodePoolID != "m0ckr" {
				t.Errorf("Case %d - Expected ID %q, got %q", i, "m0ckr", r.nodePoolID)
			}
		})
	}
}

// TestVerifyPreconditions tests cases where validating preconditions fails.
func TestVerifyPreconditions(t *testing.T) {
	var testCases = []struct {
		args         Arguments
		errorMatcher func(error) bool
	}{
		// Cluster ID is missing.
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "",
				Provider:        "aws",
			},
			errors.IsClusterNameOrIDMissingError,
		},
		// Availability zones flags are conflicting.
		{
			Arguments{
				AuthToken:             "token",
				APIEndpoint:           "https://mock-url",
				AvailabilityZonesList: []string{"fooa", "foob"},
				AvailabilityZonesNum:  3,
				ClusterNameOrID:       "cluster-id",
				Provider:              "aws",
			},
			errors.IsConflictingFlagsError,
		},
		// Availability zones number is negative on AWS.
		{
			Arguments{
				AuthToken:            "token",
				APIEndpoint:          "https://mock-url",
				AvailabilityZonesNum: -1,
				ClusterNameOrID:      "cluster-id",
				Provider:             "aws",
			},
			IsInvalidAvailabilityZones,
		},
		// Scaling min and max are not plausible.
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				ScalingMax:      3,
				ScalingMin:      5,
				Provider:        "aws",
			},
			errors.IsWorkersMinMaxInvalid,
		},
		// Using both instance type and VM size
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				InstanceType:    "something-big",
				VmSize:          "something-also-big",
				Provider:        "aws",
			},
			errors.IsConflictingFlagsError,
		},
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
				VmSize:          "something-also-big",
				ScalingMin:      3,
				ScalingMax:      1,
				Provider:        "azure",
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
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
			},
			400,
			`{"code": "INVALID_INPUT", "message": "Here is some error message"}`,
			errors.IsBadRequestError,
		},
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
			},
			403,
			`{"code": "FORBIDDEN", "message": "Here is some error message"}`,
			errors.IsAccessForbiddenError,
		},
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
			},
			404,
			`{"code": "RESOURCE_NOT_FOUND", "message": "Here is some error message"}`,
			errors.IsClusterNotFoundError,
		},
		{
			Arguments{
				AuthToken:       "token",
				APIEndpoint:     "https://mock-url",
				ClusterNameOrID: "cluster-id",
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
				if r.Method == "POST" {
					switch uri := r.URL.Path; uri {
					case "/v5/clusters/cluster-id/nodepools/":
						w.WriteHeader(tc.responseStatusCode)
						w.Write([]byte(tc.responseBody))
					default:
						t.Errorf("Unsupported route %s called in mock server", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
					}
				}
			}))
			defer mockServer.Close()

			tc.args.APIEndpoint = mockServer.URL

			clientWrapper, err := client.NewWithConfig(tc.args.APIEndpoint, tc.args.UserProvidedToken)
			if err != nil {
				t.Fatalf("Case %d - Unxpected error '%s'", i, err)
			}

			_, err = createNodePool(tc.args, tc.args.ClusterNameOrID, clientWrapper)
			if err == nil {
				t.Errorf("Case %d - Expected error, got nil", i)
			} else if !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expectec type. Got '%s'", i, err)
			}
		})
	}
}

func Test_expandAndValidateZones(t *testing.T) {
	testCases := []struct {
		name           string
		zones          []string
		provider       string
		dataCenterName string
		expectedResult []string
		errorMatcher   func(error) bool
	}{
		{
			name:           "case 0: aws zones, initials",
			zones:          []string{"a", "b", "c"},
			provider:       provider.AWS,
			dataCenterName: "eu-central-1",
			expectedResult: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			errorMatcher:   nil,
		},
		{
			name:           "case 1: aws zones, full names",
			zones:          []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			provider:       provider.AWS,
			dataCenterName: "eu-central-1",
			expectedResult: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			errorMatcher:   nil,
		},
		{
			name:           "case 2: azure zones, valid numbers",
			zones:          []string{"1", "2", "3"},
			provider:       provider.Azure,
			expectedResult: []string{"1", "2", "3"},
			errorMatcher:   nil,
		},
		{
			name:           "case 3: azure zones, not all valid numbers",
			zones:          []string{"1", "asd2", "3"},
			provider:       provider.Azure,
			expectedResult: nil,
			errorMatcher:   IsInvalidAvailabilityZones,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandAndValidateZones(tc.zones, tc.provider, tc.dataCenterName)
			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Fatalf("error not matching expected matcher, got: %s", microerror.Cause(err))
				}

				// All good. Fall through.
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if diff := cmp.Diff(result, tc.expectedResult); len(diff) > 0 {
				t.Fatalf("result not expected, got:\n %s", diff)
			}
		})
	}
}
