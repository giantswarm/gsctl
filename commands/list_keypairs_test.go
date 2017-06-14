package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/viper"
)

// TestNotLoggedIn checks whether we detect that the user is not logged in
func Test_CheckListKeypairs_NotLoggedIn(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)
	config.Config.Token = ""

	err := checkListKeypairs(ListKeypairsCommand, []string{})
	if err == nil || err.Error() != "You are not logged in. Use 'gsctl login' to log in." {
		t.Error("Unexpected error:", err)
	}
}

// TestLoggedIn checks whether we assume that the user is logged in
func Test_CheckListKeypairs_LoggedIn(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)
	config.Config.Token = "some-test-token"

	// needed to prevent search for the default cluster
	cmdClusterID = "foobar"
	err := checkListKeypairs(ListKeypairsCommand, []string{})
	if err != nil {
		t.Error("Login token not accepted:", err)
	}
}

// Not logged in, but a cluster ID is given. An error should occur.
func Test_CheckListKeypairs_NotLoggedInWithCluster(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)
	config.Config.Token = ""

	cmdClusterID = "foobar"
	err := checkListKeypairs(ListKeypairsCommand, []string{})
	if err == nil {
		t.Error("Expected error didn't occur, err was nil.")
	}
}

func Test_ListKeyPairs(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	config.Initialize(dir)

	// mock service returning key pairs. For the sake of cimplicity,
	// it doesn't care about auth tokens.
	keyPairsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
      {
        "create_date": "2017-03-17T12:41:23.053271166Z",
        "description": "Added by user marian@sendung.de using 'gsctl create kubeconfig'",
        "id": "52:64:7d:ca:75:3c:7b:46:06:2f:a0:ce:42:9a:76:c9:2b:76:aa:9e",
        "ttl_hours": 720
      },
      {
        "create_date": "2017-01-23T13:57:57.755631763Z",
        "description": "Added by user oliver.ponder@gmail.com using Happa web interface",
        "id": "74:2d:de:d2:6b:9f:4d:a5:e5:0d:eb:6e:98:14:02:6c:79:40:f6:58",
        "ttl_hours": 720
      }
    ]`))
	}))
	defer keyPairsMockServer.Close()

	cmdAPIEndpoint = keyPairsMockServer.URL
	cmdClusterID = "test-cluster-id"
	_, err := keypairsTable()
	if err != nil {
		t.Error(err)
	}
}
