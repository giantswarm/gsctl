package capabilities

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/gsctl/client"
	"github.com/google/go-cmp/cmp"
)

type testInput struct {
	Provider       string
	ReleaseVersion string
}

var capabilityTests = []struct {
	in  testInput
	out []string
}{
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "1.2.3",
		},
		[]string{},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "6.1.2",
		},
		[]string{AvailabilityZones.Name},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "6.4.0",
		},
		[]string{Autoscaling.Name, AvailabilityZones.Name},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "9.0.0",
		},
		[]string{Autoscaling.Name, AvailabilityZones.Name, NodePools.Name},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "9.1.2",
		},
		[]string{Autoscaling.Name, AvailabilityZones.Name, NodePools.Name},
	},
	{
		testInput{
			Provider:       "kvm",
			ReleaseVersion: "9.1.2",
		},
		[]string{},
	},
}

var failingCapabilityTests = []struct {
	in           testInput
	errorMatcher func(error) bool
}{
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "1.2.3.4",
		},
		IsInvalidSemVer,
	},
}

func TestGetCapabilities(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.URL.String() == "/v4/info/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"general": {
				  "provider": "aws"
				},
				"features": {
				  "nodepools": {"release_version_minimum": "9.0.0"}
				}
			  }`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "ERROR", "message": "Bad things happened"}`))
		}
	}))
	defer mockServer.Close()

	// client
	clientWrapper, err := client.NewWithConfig(mockServer.URL, "test-token")
	if err != nil {
		t.Fatalf("Error in client creation: %s", err)
	}

	for index, tt := range capabilityTests {
		service, err := New(tt.in.Provider, clientWrapper)
		if err != nil {
			t.Errorf("Test %d: Error: %s", index, err)
		}
		cap, err := service.GetCapabilities(tt.in.ReleaseVersion)
		if err != nil {
			t.Errorf("Test %d: Error: %s", index, err)
		}

		names := []string{}
		for _, capability := range cap {
			names = append(names, capability.Name)
		}
		if diff := cmp.Diff(tt.out, names, nil); diff != "" {
			t.Errorf("Test %d - Resulting args unequal. (-expected +got):\n%s", index, diff)
		}
	}
}

func TestFailingCapabilities(t *testing.T) {
	// mock info endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.URL.String() == "/v4/info/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"general": {
				  "provider": "aws"
				},
				"features": {
				  "nodepools": {"release_version_minimum": "9.0.0"}
				}
			  }`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": "ERROR", "message": "Bad things happened"}`))
		}
	}))
	defer mockServer.Close()

	// client
	clientWrapper, err := client.NewWithConfig(mockServer.URL, "test-token")
	if err != nil {
		t.Fatalf("Error in client creation: %s", err)
	}

	for index, tt := range failingCapabilityTests {
		service, err := New(tt.in.Provider, clientWrapper)
		if err != nil {
			t.Errorf("Test %d: Error: %s", index, err)
		}

		_, err = service.GetCapabilities(tt.in.ReleaseVersion)
		if err == nil {
			t.Errorf("Test %d: Expected error, got nil", index)
		}
		if !tt.errorMatcher(err) {
			t.Errorf("Test %d got different error than expectred: '%s'", index, err.Error())
		}
	}
}
