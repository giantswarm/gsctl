package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/flags"
)

func Test_UpdateOrgSetCredentials_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		url := r.URL.String()

		// GET /v4/info
		if r.Method == "GET" && url == "/v4/info/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"general": {
					"installation_name": "shire",
					"provider": "aws",
					"datacenter": "eu-central-1"
				},
				"workers": {
					"count_per_cluster": {"max": 20, "default": 3},
					"instance_type": {"options": ["m3.medium", "m3.large", "m3.xlarge"],"default": "m3.large"}
				}
			}`))
		}

		// GET /v4/organizations/
		if r.Method == "GET" && url == "/v4/organizations/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"id": "acme"}
			]`))
		}

		// POST /v4/organizations/acme/credentials/
		if r.Method == "POST" && url == "/v4/organizations/acme/credentials/" {
			w.Header().Set("Location", "/v4/organizations/acme/credentials/test/")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"code": "RESOURCE_CREATED",
				"message": "A new set of credentials has been created with ID 'test'"
			}`))
		}

	}))
	defer mockServer.Close()

	dir, err := config.TempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	args := updateOrgSetCredentialsArguments{
		apiEndpoint:     mockServer.URL,
		organizationID:  "acme",
		authToken:       "test-token",
		awsAdminRole:    "test-admin-role",
		awsOperatorRole: "test-operator-role",
	}

	flags.CmdAPIEndpoint = mockServer.URL
	InitClient()

	err = verifyUpdateOrgSetCredentialsPreconditions(args)
	if err != nil {
		t.Errorf("Verifying preconditions returned error: %s", err)
	}

	result, err := updateOrgSetCredentials(args)
	if err != nil {
		t.Error(err)
	}

	// TODO: check result content
	if result.credentialID != "test" {
		t.Errorf("Expected credential ID 'test', got %q", result.credentialID)
	}
}
