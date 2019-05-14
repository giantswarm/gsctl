package version

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_CurrentVersion(t *testing.T) {
	testVersion := "0.0.0"
	current := currentVersion()
	if current != testVersion {
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
