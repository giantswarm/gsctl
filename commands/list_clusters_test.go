package commands

import (
	"fmt"
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
		fmt.Println(r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.String() == "/v1/user/me/memberships" {
			// return organizations for the current user
			w.Write([]byte(`{
          "status_code": 10000,
          "status_text": "success",
          "data": ["acme", "foo", "giantswarm"]
        }`))
		} else {
			// return clusters for the organization
			w.Write([]byte(`{
          "status_code": 10000,
          "status_text": "success",
          "data": {
            "clusters": [
              {
                "create_date": "2017-04-16T09:30:31.192170835Z",
                "id": "` + randomClusterID() + `",
                "name": "Some random test cluster"
              }
            ]
          }
        }`))
		}
	}))
	defer mockServer.Close()

	cmdAPIEndpoint = mockServer.URL
	listClusters(ListClustersCommand, []string{})
}
