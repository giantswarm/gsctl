package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test_ListOrganizationsSuccess tests the command with inputs that should succeed.
func Test_ListOrganizationsSuccess(t *testing.T) {
	testCases := []struct {
		name         string
		jsonResponse []byte
	}{
		{
			name: "Normal output with organizations",
			jsonResponse: []byte(`[
        {"id": "acme"},
				{"id": "foo"},
				{"id": "giantswarm"}
      ]`),
		},
		{
			name:         "Empty list",
			jsonResponse: []byte(`[]`),
		},
	}

	for i, tc := range testCases {
		t.Logf("Table test case %d: %s", i, tc.name)
		orgsMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(tc.jsonResponse)
		}))
		defer orgsMockServer.Close()

		cmdAPIEndpoint = orgsMockServer.URL
		args := listOrgsArguments{
			authToken:   "some-token",
			apiEndpoint: orgsMockServer.URL,
		}
		initClient()

		err := verifyListOrgsPreconditions(args)
		if err != nil {
			t.Errorf("Table test case %d: Unexpected error in verifyListOrgsPreconditions: %#v", i, err)
		}

		_, err = orgsTable()
		if err != nil {
			t.Errorf("Table test case %d: Unexpected error in orgsTable: %#v", i, err)
		}

	}
}
