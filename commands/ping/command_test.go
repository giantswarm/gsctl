package ping

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/giantswarm/gsctl/testutils"
	"github.com/spf13/afero"
)

func Test_Command_Execution(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer mockServer.Close()

	// config
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

	err = Command.Execute()
	if err != nil {
		t.Errorf("Unexpected error %s\n", err.Error())
	}
}

func Test_Ping(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer mockServer.Close()

	duration, err := ping(Arguments{apiEndpoint: mockServer.URL})
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if duration == 0 {
		t.Error("Expected duration > 0, was 0")
	}
}

func Test_Ping_InternalServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	_, err := ping(Arguments{apiEndpoint: mockServer.URL})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Test_Ping_NonexistingEndpoint(t *testing.T) {
	_, err := ping(Arguments{apiEndpoint: "http://notexisting"})
	if err == nil {
		t.Error("Expected 'no such host' error, got <nil>")
	}
}
