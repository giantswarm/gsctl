package commands

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
        "id": "` + randomClusterID() + `",
        "name": "My dearest production cluster",
        "owner": "acme"
      },
      {
        "create_date": "2017-04-16T09:30:31.192170835Z",
        "id": "` + randomClusterID() + `",
        "name": "Abandoned cluster from the early days",
        "owner": "some_org"
      },
      {
        "create_date": "2017-10-06T02:24:55.192170835Z",
        "id": "` + randomClusterID() + `",
        "name": "A fairly recent test cluster",
        "owner": "acme"
      },
      {
        "create_date": "2017-10-10T07:24:55.192170835Z",
        "id": "` + randomClusterID() + `",
        "name": "That brand new development cluster",
        "owner": "acme_dev"
      }
    ]`))
	}))
	defer mockServer.Close()

	args := listClustersArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "testtoken",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	table, err := clustersTable(args)
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

	args := listClustersArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "testtoken",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	table, err := clustersTable(args)
	if err != nil {
		t.Error(err)
	}

	if table != "" {
		t.Errorf("Expected '', got '%s'", table)
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

	args := listClustersArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "testtoken",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	_, err = clustersTable(args)
	if !IsNotAuthorizedError(err) {
		t.Errorf("Expected notAuthorizedError, got %#v", err)
	}
}

// Test_ListClustersForbidden tests listing clusters with a 403 response.
func Test_ListClustersForbidden(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`Forbidden`))
	}))
	defer mockServer.Close()

	args := listClustersArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "testtoken",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err := verifyListClusterPreconditions(args)
	if err != nil {
		t.Error(err)
	}

	_, err = clustersTable(args)
	if !IsAccessForbiddenError(err) {
		t.Errorf("Expected accessForbiddenError, got %#v", err)
	}
}
