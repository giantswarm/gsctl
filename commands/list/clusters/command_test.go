package clusters

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/spf13/afero"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomClusterID() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 7)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Test_ListClusters tests listing a non-empty clusters table
func Test_ListClusters(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// return clusters for the organization
		w.Write([]byte(`[
			{
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"id": "fow72",
				"name": "My dearest production cluster",
				"owner": "acme",
				"path": "/v4/clusters/fow72/"
			},
			{
				"create_date": "2017-04-16T09:30:31.192170835Z",
				"id": "2sg4i",
				"name": "Abandoned cluster from the early days",
				"owner": "some_org",
				"path": "/v4/clusters/2sg4i/"
			},
			{
				"create_date": "2017-10-06T02:24:55.192170835Z",
				"id": "7ste0",
				"name": "A fairly recent test cluster",
				"owner": "acme",
				"path": "/v4/clusters/7ste0/"
			},
			{
				"create_date": "2017-10-10T07:24:55.192170835Z",
				"id": "d740d",
				"name": "That brand new development cluster",
				"owner": "acme_dev",
				"path": "/v4/clusters/d740d/"
			},
			{
				"create_date": "2017-10-10T07:24:55.192170835Z",
				"delete_date": "2019-10-10T07:24:55.192170835Z",
				"id": "del01",
				"name": "A deleted cluster",
				"owner": "acme",
				"path": "/v5/clusters/del01/"
			}
		]`))
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	args := Arguments{
		apiEndpoint:  mockServer.URL,
		authToken:    "testtoken",
		outputFormat: "table",
	}

	err = verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	table, err := getClustersOutput(args)
	if err != nil {
		t.Error(err)
	}

	t.Log(table)
}

// Test_ListClustersEmpty tests listing an empty cluster table
func Test_ListClustersEmpty(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	args := Arguments{
		apiEndpoint:  mockServer.URL,
		authToken:    "testtoken",
		outputFormat: "table",
	}

	err = verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	table, err := getClustersOutput(args)
	if err != nil {
		t.Error(err)
	}

	if table != "No clusters" {
		t.Errorf("Expected 'No clusters', got '%s'", table)
	}
}

// Test_ListClustersUnauthorized tests listing clusters with a 401 response.
func Test_ListClustersUnauthorized(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "PERMISSION_DENIED", "message": "Lorem ipsum"}`))
	}))
	defer mockServer.Close()

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Error(err)
	}

	args := Arguments{
		apiEndpoint:  mockServer.URL,
		authToken:    "testtoken",
		outputFormat: "table",
	}

	err = verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	_, err = getClustersOutput(args)
	if !errors.IsNotAuthorizedError(err) {
		t.Errorf("Expected NotAuthorizedError, got %#v", err)
	}
}
