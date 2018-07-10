package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	gsclient "github.com/giantswarm/gsclientgen/client"
	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/models"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	rootcerts "github.com/hashicorp/go-rootcerts"
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
	apiClient, clientErr := New(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	_, _, err := apiClient.GetUserOrganizations("foo")
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
	apiClient, clientErr := New(clientConfig)
	if clientErr != nil {
		t.Error(clientErr)
	}
	// We use GetUserOrganizations just to issue a request. We could use any other
	// API call, it wouldn't matter.
	_, apiResponse, _ := apiClient.GetUserOrganizations("foo")

	gr, err := ParseGenericResponse(apiResponse.Payload)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(gr.Message, clientConfig.UserAgent) {
		t.Error("UserAgent string could not be found")
	}
}

// TestRedactPasswordArgs tests redactPasswordArgs()
func TestRedactPasswordArgs(t *testing.T) {
	argtests := []struct {
		in  string
		out string
	}{
		// these remain unchangd
		{"foo", "foo"},
		{"foo bar", "foo bar"},
		{"foo bar blah", "foo bar blah"},
		{"foo bar blah -p mypass", "foo bar blah -p mypass"},
		{"foo bar blah -p=mypass", "foo bar blah -p=mypass"},
		// these will be altered
		{"foo bar blah --password mypass", "foo bar blah --password REDACTED"},
		{"foo bar blah --password=mypass", "foo bar blah --password=REDACTED"},
		{"foo login blah -p mypass", "foo login blah -p REDACTED"},
		{"foo login blah -p=mypass", "foo login blah -p=REDACTED"},
	}

	for _, tt := range argtests {
		in := strings.Split(tt.in, " ")
		out := strings.Join(redactPasswordArgs(in), " ")
		if out != tt.out {
			t.Errorf("want '%q', have '%s'", tt.in, tt.out)
		}
	}
}

// TestClientV2CreateAuthToken checks out how creating an auth token works in
// our new client
func TestClientV2CreateAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth_token": "e5239484-2299-41df-b901-d0568db7e3f9"}`))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		t.Error(rootCertsErr)
	}

	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	gsClient := gsclient.New(transport, strfmt.Default)

	params := auth_tokens.NewCreateAuthTokenParams()
	params.SetBody(&models.V4CreateAuthTokenRequest{
		Email:          "foo",
		PasswordBase64: "bar",
	})

	// we use nil as runtime.ClientAuthInfoWriter here
	response, err := gsClient.AuthTokens.CreateAuthToken(params, nil)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%#v", response.Payload)

	if response.Payload.AuthToken != "e5239484-2299-41df-b901-d0568db7e3f9" {
		t.Errorf("Didn't get the expected token. Got %s", response.Payload.AuthToken)
	}
}

// TestClientV2DeleteAuthToken checks out how to issue an authenticted request
// using the new client
func TestClientV2DeleteAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "giantswarm test-token" {
			t.Error("Bad authorization header:", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been succesfully deleted."}`))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		t.Error(rootCertsErr)
	}

	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	gsClient := gsclient.New(transport, strfmt.Default)

	params := auth_tokens.NewDeleteAuthTokenParams().WithAuthorization("giantswarm test-token")

	response, err := gsClient.AuthTokens.DeleteAuthToken(params, nil)
	if err != nil {
		t.Error(err)
	}

	if response.Payload.Code != "RESOURCE_DELETED" {
		t.Errorf("Didn't get the RESOURCE_DELETED message. Got '%s'", response.Payload.Code)
	}
}
