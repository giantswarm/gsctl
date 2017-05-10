package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Ping(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer mockServer.Close()

	duration, err := ping(mockServer.URL)
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

	_, err := ping(mockServer.URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Test_Ping_NonexistingEndpoint(t *testing.T) {
	_, err := ping("http://not.existing")
	if err == nil {
		t.Error("Expected 'no such host' error, got nil")
	}
}
