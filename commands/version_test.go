package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-semver/semver"
)

// Test_LatestVersion checks the basic functions of latestVersion()
func Test_LatestVersion(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("0.6.1\n"))
	}))
	defer mockServer.Close()

	version, err := latestVersion(mockServer.URL)
	if err != nil {
		t.Error(err)
	}

	testVersion := semver.New("0.6.1")
	if !version.Equal(*testVersion) {
		t.Error("Version equality check failed.")
	}
}

func Test_CurrentVersion(t *testing.T) {
	testVersion := semver.New("0.0.0")
	current := currentVersion()
	if !current.Equal(*testVersion) {
		t.Error("Version equality check failed.")
	}
}

func Test_CheckUpdateAvailable(t *testing.T) {
	latestPath := "/giantswarm/gsctl/releases/latest"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == latestPath {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Location", "https://github.com/giantswarm/gsctl/releases/tag/1.2.3")
			w.WriteHeader(http.StatusFound)
			w.Write([]byte("<html><body>You are being redirected</body></html>"))
		} else {
			t.Errorf("Unhandled URL called: %s", r.URL.String())
		}
	}))
	defer mockServer.Close()

	info, err := checkUpdateAvailable(mockServer.URL + latestPath)
	if err != nil {
		t.Error(err)
	}

	if !info.updateAvailable {
		t.Error("checkUpdateAvailable didn't produce the expected conclusion.")
	}
}

func Test_VersionCheckDue(t *testing.T) {
	if !versionCheckDue() {
		t.Error("versionCheckDue() should return true, returned false")
	}
}
