package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Test_ListReleases_Empty simulates the situation where there are no releases
// (which is an exception)
func Test_ListReleases_Empty(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// mock service returning empty release array.
	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer releasesMockServer.Close()

	// needed to prevent search for the default cluster
	args := listReleasesArguments{
		apiEndpoint: releasesMockServer.URL,
		token:       "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err = listReleasesPreconditions(&args)
	if err != nil {
		t.Error(err)
	}

	result, listErr := listReleases(args)
	if listErr != nil {
		t.Error(listErr)
	}
	if len(result.releases) > 0 {
		t.Error("Got releases where we expected none.")
	}
}

// Test_ListReleases_Connection_Unavailable simulates the situation where we
// cannot reach the endpoint
func Test_ListReleases_Connection_Unavailable(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// needed to prevent search for the default cluster
	args := listReleasesArguments{
		apiEndpoint: "http://localhost:45454",
		token:       "my-token",
	}

	err = listReleasesPreconditions(&args)
	if err != nil {
		t.Error(err)
	}

	_, listErr := listReleases(args)
	if !IsNoResponseError(listErr) {
		t.Errorf("Expected noResponseError, got '%s'", listErr)
	}
}

// Test_ListReleases_NotFound simulates the situation where the cluster
// to list releases for is not found
func Test_ListReleases_NotFound(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	releasesMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "The cluster could not be found."}`))
	}))
	defer releasesMockServer.Close()

	args := listReleasesArguments{
		apiEndpoint: releasesMockServer.URL,
		token:       "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err = listReleasesPreconditions(&args)
	if err != nil {
		t.Error(err)
	}

	_, listErr := listReleases(args)
	if listErr == nil {
		t.Error("No error occurred where we expected one.")
	} else if !IsClusterNotFoundError(listErr) {
		t.Errorf("Expected error '%s', got '%s'.", clusterNotFoundError, listErr)
	}
}

// Test_ListReleases_Nonempty simulates listing releases where several
// items are returned.
func Test_ListReleases_Nonempty(t *testing.T) {
	dir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// mock service returning releases. For the sake of simplicity,
	// it doesn't care about auth tokens.
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

	args := listReleasesArguments{
		apiEndpoint: releasesMockServer.URL,
		token:       "my-token",
	}

	cmdAPIEndpoint = releasesMockServer.URL
	initClient()

	err = listReleasesPreconditions(&args)
	if err != nil {
		t.Error(err)
	}

	result, listErr := listReleases(args)
	if listErr != nil {
		t.Error(listErr)
	}

	if len(result.releases) != 2 {
		t.Errorf("We expected 2 releases, got %d", len(result.releases))
	}

	if *result.releases[0].Version != "0.10.0" || *result.releases[1].Version != "0.1.0" {
		t.Error("Releases returned were not in the expected order.")
	}
}
