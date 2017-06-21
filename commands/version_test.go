package commands

import (
	"fmt"
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
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("0.1.0\n"))
	}))
	defer mockServer.Close()

	info, err := checkUpdateAvailable(mockServer.URL)
	if err != nil {
		t.Error(err)
	}

	if !info.updateAvailable {
		t.Error("checkUpdateAvailable didn't produce the expected conclusion.")
	}

	fmt.Println(updateInfo(info))
}

func Test_VersionCheckDue(t *testing.T) {
	if !versionCheckDue() {
		t.Error("versionCheckDue() should return true, returned false")
	}
}
