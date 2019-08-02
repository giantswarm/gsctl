package setcredentials

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/testutils"
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

	configYAML := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, configYAML)
	if err != nil {
		t.Error(err)
	}

	args := collectArguments()
	args.organizationID = "acme"
	args.awsAdminRole = "test-admin-role"
	args.awsOperatorRole = "test-operator-role"

	err = verifyPreconditions(args)
	if err != nil {
		t.Errorf("Verifying preconditions returned error: %s", err)
	}

	result, err := setOrgCredentials(args)
	if err != nil {
		t.Error(err)
	}

	if result.credentialID != "test" {
		t.Errorf("Expected credential ID 'test', got %q", result.credentialID)
	}

	expected := "Credentials set successfully"
	output := testutils.CaptureOutput(func() {
		printResult(Command, []string{""})
	})
	if !strings.Contains(output, expected) {
		t.Logf("Command output: %q", output)
		t.Errorf("Command output did not contain expected part %q\n", expected)
	}
}
