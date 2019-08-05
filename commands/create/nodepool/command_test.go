package nodepool

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

// TODO:
// cases which should succeed:
//   - setting only scaling min or only max should result in proper setting

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
			[]string{"cluster-id"},
			func() {
				Command.ParseFlags([]string{"cluster-id", "--name", "my-name"})
			},
			Arguments{
				APIEndpoint: "https://endpoint",
				AuthToken:   "some-token",
				ClusterID:   "cluster-id",
				Name:        "my-name",
				Scheme:      "giantswarm",
			},
		},
		{
			[]string{"cluster-id"},
			func() {
				Command.ParseFlags([]string{
					"cluster-id",
					"--num-availability-zones=3",
					"--aws-instance-type=instance-type",
					"--nodes-min=5",
					"--nodes-max=10",
				})
			},
			Arguments{
				APIEndpoint:          "https://endpoint",
				AuthToken:            "some-token",
				ClusterID:            "cluster-id",
				Name:                 "my-name",
				Scheme:               "giantswarm",
				AvailabilityZonesNum: 3,
				InstanceType:         "instance-type",
				ScalingMin:           5,
				ScalingMax:           10,
			},
		},
	}

	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  https://endpoint:
    email: email@example.com
    token: some-token
selected_endpoint: https://endpoint`

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
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

// TestSuccess tests node pool creation with cases that are expected to succeed.
func TestSuccess(t *testing.T) {
	var testCases = []struct {
		args         Arguments
		responseBody string
	}{
		// Minimal node pool creation.
		{
			Arguments{
				ClusterID: "cluster-id",
				AuthToken: "token",
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

			r, err := createNodePool(tc.args)
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
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				ClusterID:   "",
			},
			errors.IsClusterIDMissingError,
		},
		// Availability zones flags are conflicting.
		{
			Arguments{
				AuthToken:             "token",
				APIEndpoint:           "https://mock-url",
				AvailabilityZonesList: []string{"fooa", "foob"},
				AvailabilityZonesNum:  3,
				ClusterID:             "cluster-id",
			},
			errors.IsConflictingFlagsError,
		},
		// Scaling min and max are not plausible.
		{
			Arguments{
				AuthToken:   "token",
				APIEndpoint: "https://mock-url",
				ClusterID:   "cluster-id",
				ScalingMax:  3,
				ScalingMin:  5,
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
