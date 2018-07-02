package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ListOrganizations(t *testing.T) {
	// mock service returning organizations the user is member of
	orgsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
        {"id": "acme"},
				{"id": "foo"},
				{"id": "giantswarm"}
      ]`))
	}))
	defer orgsMockServer.Close()

	cmdAPIEndpoint = orgsMockServer.URL
	initClient()
	_, err := orgsTable()
	if err != nil {
		t.Error(err)
	}

	listOrgs(ListOrgsCommand, []string{})
}

func Test_ListOrganizationsEmpty(t *testing.T) {
	// mock service returning key pairs
	orgsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer orgsMockServer.Close()

	cmdAPIEndpoint = orgsMockServer.URL
	initClient()
	_, err := orgsTable()
	if err != nil {
		t.Error(err)
	}

	listOrgs(ListOrgsCommand, []string{})
}
