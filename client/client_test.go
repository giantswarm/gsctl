package client

import (
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	clientConfig := Configuration{
		Endpoint: "http://httpbin.org/delay/2?",
		Timeout:  1 * time.Second,
	}
	apiClient := NewClient(clientConfig)
	_, _, err := apiClient.GetUserOrganizations("foo", "foo", "foo", "foo")
	if err == nil || !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Error("Expected Client.Timeout error, got:", err)
	}
}

func TestUserAgent(t *testing.T) {
	clientConfig := Configuration{
		Endpoint:  "https://httpbin.org/user-agent?",
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
