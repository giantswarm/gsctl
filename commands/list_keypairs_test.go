package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/viper"
)

// Test_ListKeypairs_NotLoggedIn checks whether we are detecting whether or not the user
// is logged in or provides an auth token.
func Test_ListKeypairs_NotLoggedIn(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)
	config.Config.Token = ""

	args := listKeypairsArguments{}
	err := listKeypairsValidate(&args)
	if err == nil {
		t.Error("No error thrown where we expected an error.")
	}
}

// Test_ListKeypairs_Empty simulates the situation where there are no
// key pairs for a given cluster.
func Test_ListKeypairs_Empty(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	// mock service returning empty key pair array.
	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer keyPairsMockServer.Close()

	// needed to prevent search for the default cluster
	args := listKeypairsArguments{}
	args.apiEndpoint = keyPairsMockServer.URL
	args.token = "my-token"
	args.clusterID = "my-cluster"

	err := listKeypairsValidate(&args)
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
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "The cluster could not be found."}`))
	}))
	defer keyPairsMockServer.Close()

	args := listKeypairsArguments{}
	args.apiEndpoint = keyPairsMockServer.URL
	args.token = "my-token"
	args.clusterID = "unknown-cluster"

	err := listKeypairsValidate(&args)
	if err != nil {
		t.Error(err)
	}

	_, listErr := listKeypairs(args)
	if listErr == nil {
		t.Error("No error occurred where we expected one.")
	} else if !IsClusterNotFoundError(listErr) {
		t.Errorf("Expected error '%s', got '%s'.", clusterNotFoundError, listErr)
	}
}

// Test_ListKeyPairs_Nonempty simulates listing key pairs where several
// items are returned.
func Test_ListKeyPairs_Nonempty(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

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

	args := listKeypairsArguments{}
	args.apiEndpoint = keyPairsMockServer.URL
	args.token = "my-token"
	args.clusterID = "my-cluster"

	err := listKeypairsValidate(&args)
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
	if result.keypairs[1].Id != "52:64:7d:ca:75:3c:7b:46:06:2f:a0:ce:42:9a:76:c9:2b:76:aa:9e" {
		t.Error("Keypairs returned were not in the expected order.")
	}
}
