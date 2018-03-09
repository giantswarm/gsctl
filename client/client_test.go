package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestTimeout tests that we return the correct error type in case of a client timeout
func TestTimeout(t *testing.T) {
	clientTimeout := 1 * time.Second

	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce a timeout longer than the client's
		time.Sleep(2 * clientTimeout)
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	apiClient, clientErr := NewClient(Configuration{
		Endpoint: ts.URL,
		Timeout:  clientTimeout,
	})
	if clientErr != nil {
		t.Error(clientErr)
	}

	authHeader := "giantswarm test-token"

	_, _, err := apiClient.DefaultApi.GetUserOrganizations(context.TODO(), authHeader, nil)
	if err == nil {
		t.Error("Expected Timeout error, got nil")
	} else {
		if !IsTimeoutError(err) {
			t.Error("Expected Timeout error, got", err)
		}
	}
}

// TestUserAgent tests whether request have the proper User-Agent header
func TestUserAgent(t *testing.T) {
	userAgent := "my own user agent/1.0"

	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return valid JSON containing user agent string received
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("Expected %q, got %q", userAgent, r.Header.Get("User-Agent"))
		}
		w.Write([]byte(`{
		  "general": {
		    "installation_name": "shire",
		    "provider": "aws",
		    "datacenter": "eu-central-1"
		  },
		  "workers": {
		    "count_per_cluster": {
		      "max": 20,
		      "default": 3
		    },
		    "instance_type": {
		      "options": [
		        "m3.medium", "m3.large", "m3.xlarge"
		      ],
		      "default": "m3.large"
		    }
		  }
		}`))
	}))
	defer ts.Close()

	apiClient, clientErr := NewClient(Configuration{
		Endpoint:  ts.URL,
		UserAgent: userAgent,
	})
	if clientErr != nil {
		t.Error(clientErr)
	}

	authHeader := "giantswarm test-token"

	// We use GetInfo as an example API request
	_, _, callErr := apiClient.DefaultApi.GetInfo(context.TODO(), authHeader, nil)
	if callErr != nil {
		t.Error(callErr)
	}
}

// TestAuth tests if we can pass an authorization token
func TestAuth(t *testing.T) {
	token := "test-token"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Authorization") != "test-token" {
			t.Errorf("Expected auth token %q, got %q", token, r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": "BAD_REQUEST", "message": "user-agent: ` + r.Header.Get("Authorization") + `"}`))
	}))
	defer mockServer.Close()

	// TODO
}

// TestParseGenericResponse tests if we can parse a generic error message JSON body
func TestParseGenericResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{
		  "code": "INTERNAL_ERROR",
			"message": "Some random message"
		}`))
	}))
	defer mockServer.Close()

	apiClient, clientErr := NewClient(Configuration{Endpoint: mockServer.URL})
	if clientErr != nil {
		t.Error(clientErr)
	}

	authHeader := "giantswarm test-token"

	// We use GetInfo as an example API request
	_, resp, callErr := apiClient.DefaultApi.GetInfo(context.TODO(), authHeader, nil)
	if callErr == nil {
		t.Error("Expected HTTP error, got nil")
	} else {
		t.Logf("callErr: %+v", callErr)
		t.Logf("resp: %+v", resp)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err.Error())
		}

		_, parseErr := ParseGenericResponse(body)
		if parseErr != nil {
			t.Error(parseErr)
		}
	}
}
