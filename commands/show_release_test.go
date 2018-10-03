package commands

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// TestShowRelease tests fetching release details
func TestShowRelease(t *testing.T) {
	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
		  {
				"timestamp": "2017-10-15T12:00:00Z",
		    "version": "0.1.0",
		    "active": true,
		    "changelog": [
		      {
		        "component": "vault",
		        "description": "Vault version updated."
		      },
		      {
		        "component": "flannel",
		        "description": "Flannel version updated."
		      }
		    ],
		    "components": [
		      {
		        "name": "vault",
		        "version": "0.7.2"
		      },
		      {
		        "name": "flannel",
		        "version": "0.8.0"
		      },
		      {
		        "name": "calico",
		        "version": "2.6.1"
		      },
		      {
		        "name": "docker",
		        "version": "1.12.5"
		      },
		      {
		        "name": "etcd",
		        "version": "3.2.2"
		      },
		      {
		        "name": "kubedns",
		        "version": "1.14.4"
		      },
		      {
		        "name": "kubernetes",
		        "version": "1.8.0"
		      },
		      {
		        "name": "nginx-ingress-controller",
		        "version": "0.8.0"
		      }
		    ]
		  },
			{
				"timestamp": "2017-10-27T16:21:00Z",
		    "version": "0.10.0",
		    "active": true,
		    "changelog": [
		      {
		        "component": "vault",
		        "description": "Vault version updated."
		      },
		      {
		        "component": "flannel",
		        "description": "Flannel version updated."
		      },
		      {
		        "component": "calico",
		        "description": "Calico version updated."
		      },
		      {
		        "component": "docker",
		        "description": "Docker version updated."
		      },
		      {
		        "component": "etcd",
		        "description": "Etcd version updated."
		      },
		      {
		        "component": "kubedns",
		        "description": "KubeDNS version updated."
		      },
		      {
		        "component": "kubernetes",
		        "description": "Kubernetes version updated."
		      },
		      {
		        "component": "nginx-ingress-controller",
		        "description": "Nginx-ingress-controller version updated."
		      }
		    ],
		    "components": [
		      {
		        "name": "vault",
		        "version": "0.7.3"
		      },
		      {
		        "name": "flannel",
		        "version": "0.9.0"
		      },
		      {
		        "name": "calico",
		        "version": "2.6.2"
		      },
		      {
		        "name": "docker",
		        "version": "1.12.6"
		      },
		      {
		        "name": "etcd",
		        "version": "3.2.7"
		      },
		      {
		        "name": "kubedns",
		        "version": "1.14.5"
		      },
		      {
		        "name": "kubernetes",
		        "version": "1.8.1"
		      },
		      {
		        "name": "nginx-ingress-controller",
		        "version": "0.9.0"
		      }
		    ]
		  }
		]`))
	}))
	defer releasesMockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    releasesMockServer.URL,
		releaseVersion: "0.10.0",
		scheme:         "giantswarm",
		authToken:      "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err := verifyShowReleasePreconditions(testArgs, []string{testArgs.releaseVersion})
	if err != nil {
		t.Error(err)
	}

	details, showErr := getReleaseDetails(testArgs.releaseVersion,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)
	if showErr != nil {
		t.Error(showErr)
	}

	if *details.Version != testArgs.releaseVersion {
		t.Errorf("Expected release version '%s', got '%s'", testArgs.releaseVersion, *details.Version)
	}

}

// TestShowReleaseNotAuthorized tests HTTP 401 error handling
func TestShowReleaseNotAuthorized(t *testing.T) {
	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"code":"INVALID_CREDENTIALS",
				"message":"The requested resource cannot be accessed using the provided credentials. (token not found: unauthenticated)"
			}`))
		}
	}))
	defer releasesMockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    releasesMockServer.URL,
		releaseVersion: "0.10.0",
		scheme:         "giantswarm",
		authToken:      "my-wrong-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err := verifyShowReleasePreconditions(testArgs, []string{testArgs.releaseVersion})
	if err != nil {
		t.Error(err)
	}

	_, err = getReleaseDetails(testArgs.releaseVersion,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected notAuthorizedError, got nil")
	}

	if !IsNotAuthorizedError(err) {
		t.Errorf("Expected notAuthorizedError, got '%s'", err.Error())
	}
}

// TestShowReleaseNotFound tests HTTP 404 error handling
func TestShowReleaseNotFound(t *testing.T) {
	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
		  {
			"timestamp": "2017-10-27T16:21:00Z",
		    "version": "0.10.0",
		    "active": true,
		    "changelog": [
		   	  {
		        "component": "nginx-ingress-controller",
		        "description": "Nginx-ingress-controller version updated."
		      }
		    ],
		    "components": [
		      {
		        "name": "kubernetes",
		        "version": "1.8.1"
		      }
		    ]
		  }
		]`))
	}))
	defer releasesMockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    releasesMockServer.URL,
		releaseVersion: "non-existing-release-version",
		scheme:         "giantswarm",
		authToken:      "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err := verifyShowReleasePreconditions(testArgs, []string{testArgs.releaseVersion})
	if err != nil {
		t.Error(err)
	}

	_, err = getReleaseDetails(testArgs.releaseVersion,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected releaseNotFoundError, got nil")
	}

	if !IsReleaseNotFoundError(err) {
		t.Errorf("Expected releaseNotFoundError, got '%s'", err.Error())
	}
}

// TestShowReleaseInternalServerError tests HTTP 500 error handling
func TestShowReleaseInternalServerError(t *testing.T) {
	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if r.Method == "GET" {
			w.Write([]byte(`{
				"code":"INTERNAL_ERROR",
				"message":"An unexpected error occurred. Sorry for the inconvenience."
		}`))
		}
	}))
	defer releasesMockServer.Close()

	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    releasesMockServer.URL,
		releaseVersion: "non-existing-release-version",
		scheme:         "giantswarm",
		authToken:      "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err := verifyShowReleasePreconditions(testArgs, []string{testArgs.releaseVersion})
	if err != nil {
		t.Error(err)
	}

	_, err = getReleaseDetails(testArgs.releaseVersion,
		testArgs.scheme, testArgs.authToken, testArgs.apiEndpoint)

	if err == nil {
		t.Fatal("Expected internalServerError, got nil")
	}

	if !IsInternalServerError(err) {
		t.Errorf("Expected internalServerError, got '%s'", err.Error())
	}
}

// TestShowReleaseNotLoggedIn tests the case where the client is not logged in
func TestShowReleaseNotLoggedIn(t *testing.T) {
	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    "foo.bar",
		releaseVersion: "release-version",
		authToken:      "",
	}

	err := verifyShowReleasePreconditions(testArgs, []string{testArgs.releaseVersion})
	if !IsNotLoggedInError(err) {
		t.Errorf("Expected notLoggedInError, got '%s'", err.Error())
	}

}

// TestShowReleaseMissingID tests the case where the release version is missing
func TestShowReleaseMissingID(t *testing.T) {
	// temp config
	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)

	testArgs := showReleaseArguments{
		apiEndpoint:    "foo.bar",
		releaseVersion: "",
		authToken:      "auth-token",
	}

	err := verifyShowReleasePreconditions(testArgs, []string{})
	if !IsReleaseVersionMissingError(err) {
		t.Errorf("Expected releaseVersionMissingError, got '%s'", err.Error())
	}

}
