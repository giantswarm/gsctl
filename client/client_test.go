package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce a timeout longer than the client's
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	clientConfig := Configuration{
		Endpoint: ts.URL,
		Timeout:  1 * time.Second,
	}
	apiClient := NewClient(clientConfig)
	_, _, err := apiClient.GetUserOrganizations("foo", "foo", "foo", "foo")
	if err == nil || !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Error("Expected Client.Timeout error, got:", err)
	}
}

func TestUserAgent(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return valid JSON containing user agent string received
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"user-agent": "` + r.Header.Get("User-Agent") + `"}`))
	}))
	defer ts.Close()

	clientConfig := Configuration{
		Endpoint:  ts.URL,
		UserAgent: "my own user agent/1.0",
	}
	apiClient := NewClient(clientConfig)
	_, apiResponse, err := apiClient.GetUserOrganizations("foo", "foo", "foo", "foo")
	if err != nil {
		t.Error(err)
	}
	responseBody := string(apiResponse.Payload)
	if !strings.Contains(responseBody, clientConfig.UserAgent) {
		t.Error("UserAgent string could not be found")
	}
}
