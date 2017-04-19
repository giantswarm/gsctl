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
		w.Write([]byte(`{
        "status_code": 10000,
        "status_text": "success",
        "data": ["acme", "foo", "giantswarm"]
      }`))
	}))
	defer orgsMockServer.Close()

	cmdAPIEndpoint = orgsMockServer.URL
	_, err := orgsTable()
	if err != nil {
		t.Error(err)
	}

	listOrgs(ListOrgsCommand, []string{})
}

func Test_ListOrganizationsEmpty(t *testing.T) {
	// mock service returning key-pairs
	orgsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
        "status_code": 10000,
        "status_text": "success",
        "data": []
      }`))
	}))
	defer orgsMockServer.Close()

	cmdAPIEndpoint = orgsMockServer.URL
	_, err := orgsTable()
	if err != nil {
		t.Error(err)
	}

	listOrgs(ListOrgsCommand, []string{})
}
