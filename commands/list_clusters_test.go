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

	cmdAPIEndpoint = mockServer.URL
	checkListClusters(ListClustersCommand, []string{})

	table, err := clustersTable()
	if err != nil {
		t.Error(err)
	}

	t.Log(table)
	listClusters(ListClustersCommand, []string{})
}

func Test_ListClustersEmpty(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	cmdAPIEndpoint = mockServer.URL
	checkListClusters(ListClustersCommand, []string{})

	table, err := clustersTable()
	if err != nil {
		t.Error(err)
	}

	t.Log(table)
	listClusters(ListClustersCommand, []string{})
}
