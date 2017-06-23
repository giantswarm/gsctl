package client

import (
	"fmt"
	"net"
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
	apiClient, clientErr := NewClient(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	_, _, err := apiClient.GetUserOrganizations("foo", "foo", "foo", "foo")
	if err == nil {
		t.Error("Expected Timeout error, got nil")
	} else {
		if err, ok := err.(net.Error); ok && !err.Timeout() {
			t.Error("Expected Timeout error, got", err)
		}
	}
}

// TestUserAgent tests whether request have the proper User-Agent header
// and if ParseGenericResponse works as expected
func TestUserAgent(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return valid JSON containing user agent string received
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": "BAD_REQUEST", "message": "user-agent: ` + r.Header.Get("User-Agent") + `"}`))
	}))
	defer ts.Close()

	clientConfig := Configuration{
		Endpoint:  ts.URL,
		UserAgent: "my own user agent/1.0",
	}
	apiClient, clientErr := NewClient(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	// We use GetUserOrganizations just to issue a request. We could use any other
	// API call, it wouldn't matter.
	_, apiResponse, _ := apiClient.GetUserOrganizations("foo", "foo", "foo", "foo")

	gr, err := ParseGenericResponse(apiResponse.Payload)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(gr.Message, clientConfig.UserAgent) {
		t.Error("UserAgent string could not be found")
	}
}
