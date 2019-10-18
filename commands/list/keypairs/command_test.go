package keypairs

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

// Test_ListKeypairs_NotLoggedIn checks whether we are detecting whether or not the user
// is logged in or provides an auth token. Here we use an empty config to have
// an unauthorized user.
func Test_ListKeypairs_NotLoggedIn(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	args := Arguments{}

	err = listKeypairsValidate(&args)
	if err == nil {
		t.Error("No error thrown where we expected an error.")
	}
}

// Test_ListKeypairs_Empty simulates the situation where there are no
// key pairs for a given cluster.
func Test_ListKeypairs_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	// mock service returning empty key pair array.
	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer keyPairsMockServer.Close()

	// needed to prevent search for the default cluster
	args := Arguments{
		apiEndpoint:  keyPairsMockServer.URL,
		clusterID:    "my-cluster",
		outputFormat: "table",
		token:        "my-token",
	}

	err = listKeypairsValidate(&args)
	if err != nil {
		t.Error(err)
	}

	result, listErr := listKeypairs(args)
	if listErr != nil {
		t.Error(listErr)
	}
	if len(result.keypairs) > 0 {
		t.Error("Got key pairs where we expected none.")
	}
}

// Test_ListKeypairs_NotFound simulates the situation where the cluster
// to list key pairs for is not found
func Test_ListKeypairs_NotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "The cluster could not be found."}`))
	}))
	defer keyPairsMockServer.Close()

	args := Arguments{
		apiEndpoint:  keyPairsMockServer.URL,
		clusterID:    "unknown-cluster",
		outputFormat: "json",
		token:        "my-token",
	}

	err = listKeypairsValidate(&args)
	if err != nil {
		t.Error(err)
	}

	_, listErr := listKeypairs(args)
	if listErr == nil {
		t.Error("No error occurred where we expected one.")
	} else if !errors.IsClusterNotFoundError(listErr) {
		t.Errorf("Expected error '%s', got '%s'.", errors.ClusterNotFoundError, listErr)
	}
}

// Test_ListKeyPairs_Nonempty simulates listing key pairs where several
// items are returned.
func Test_ListKeyPairs_Nonempty(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	// mock service returning key pairs. For the sake of simplicity,
	// it doesn't care about auth tokens.
	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"create_date": "2017-01-23T13:57:57.755631763Z",
				"description": "Added by user oliver.ponder@gmail.com using Happa web interface",
				"id": "74:2d:de:d2:6b:9f:4d:a5:e5:0d:eb:6e:98:14:02:6c:79:40:f6:58",
				"ttl_hours": 720
			},
			{
				"create_date": "2017-03-17T12:41:23.053271166Z",
				"description": "Added by user marian@sendung.de using 'gsctl create kubeconfig'",
				"id": "52:64:7d:ca:75:3c:7b:46:06:2f:a0:ce:42:9a:76:c9:2b:76:aa:9e",
				"ttl_hours": 720
			}
		]`))
	}))
	defer keyPairsMockServer.Close()

	args := Arguments{
		apiEndpoint:  keyPairsMockServer.URL,
		clusterID:    "my-cluster",
		outputFormat: "json",
		token:        "my-token",
	}

	err = listKeypairsValidate(&args)
	if err != nil {
		t.Error(err)
	}

	result, listErr := listKeypairs(args)
	if listErr != nil {
		t.Error(listErr)
	}
	if len(result.keypairs) != 2 {
		t.Errorf("We expected %d key pairs, got %d", 2, len(result.keypairs))
	}
	if result.keypairs[1].ID != "52:64:7d:ca:75:3c:7b:46:06:2f:a0:ce:42:9a:76:c9:2b:76:aa:9e" {
		t.Error("Keypairs returned were not in the expected order.")
	}
}

func Test_ListKeyPairsOutput(t *testing.T) {
	jsonOutput := `[
  {
    "create_date": "2017-01-23T13:57:57.755631763Z",
    "description": "Added by user oliver.ponder@gmail.com using Happa web interface",
    "id": "74:2d:de:d2:6b:9f:4d:a5:e5:0d:eb:6e:98:14:02:6c:79:40:f6:58",
    "ttl_hours": 720
  },
  {
    "create_date": "2017-03-17T12:41:23.053271166Z",
    "description": "Added by user marian@sendung.de using 'gsctl create kubeconfig'",
    "id": "52:64:7d:ca:75:3c:7b:46:06:2f:a0:ce:42:9a:76:c9:2b:76:aa:9e",
    "ttl_hours": 720
  }
]
`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("%s %s", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonOutput))
	}))
	defer mockServer.Close()

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "case 0: table output",
			args:           []string{"-c=foo"},
			expectedOutput: "CREATED                 EXPIRES                 ID          DESCRIPTION                                                      CN  O\n2017 Jan 23, 13:57 UTC  2017 Feb 22, 13:57 UTC  742dded26…  Added by user oliver.ponder@gmail.com using Happa web interface      \n2017 Mar 17, 12:41 UTC  2017 Apr 16, 12:41 UTC  52647dca7…  Added by user marian@sendung.de using 'gsctl create kubeconfig'      \n",
		},

		{
			name:           "case 1: json output",
			args:           []string{"-c=foo", "-o=json"},
			expectedOutput: jsonOutput,
		},
	}

	// temp config
	// config
	configYAML := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Error(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := testutils.CaptureOutput(func() {
				initFlags()
				Command.SetArgs(tc.args)
				Command.Execute()
			})

			if !cmp.Equal(output, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, output))
			}
		})
	}
}
