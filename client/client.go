package client

import (
	"crypto/tls"
	"encoding/base64"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	gsclient "github.com/giantswarm/gsclientgen/client"
	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	rootcerts "github.com/hashicorp/go-rootcerts"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified
	DefaultTimeout = 60 * time.Second

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	requestIDHeader = randomRequestID()
	cmdLine = getCommandLine()
}

// Configuration is the configuration to be used both by the latest as well
// as the old client based on gsclientgen.v1
type Configuration struct {
	// AuthHeader is the header we should use to make API calls
	AuthHeader string

	// Endpoint is the base URL of the API
	Endpoint string

	// Timeout is the maximum time to wait for API requests to succeed
	Timeout time.Duration

	// UserAgent identifier
	UserAgent string
}

// WrapperV2 is the structure holding representing our latest API client
type WrapperV2 struct {
	gsclient *gsclient.Gsclientgen
}

// NewV2 creates a client based on the latest gsclientgen version
func NewV2(conf *Configuration) (*WrapperV2, error) {
	if conf.Endpoint == "" {
		return nil, microerror.Mask(endpointNotSpecifiedError)
	}

	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, microerror.Mask(endpointInvalidError)
	}

	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		return nil, microerror.Mask(rootCertsErr)
	}

	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}
	transport.Transport = setUserAgent(transport.Transport, conf.UserAgent)

	return &WrapperV2{
		gsclient: gsclient.New(transport, strfmt.Default),
	}, nil
}

type roundTripperWithUserAgent struct {
	inner http.RoundTripper
	Agent string
}

// RoundTrip overwrites the http.RoundTripper.RoundTrip function to add our
// User-Agent HTTP header to a request
func (rtwua *roundTripperWithUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", rtwua.Agent)
	return rtwua.inner.RoundTrip(r)
}

// setUserAgent sets the User-Agent header value for subsequent requests
// made using this transport
func setUserAgent(inner http.RoundTripper, userAgent string) http.RoundTripper {
	return &roundTripperWithUserAgent{
		inner: inner,
		Agent: userAgent,
	}
}

// randomRequestID returns a new request ID
func randomRequestID() string {
	size := 14
	b := make([]rune, size)
	for i := range b {
		b[i] = randomStringCharset[rand.Intn(len(randomStringCharset))]
	}
	return string(b)
}

// getCommandLine returns the command line that has been called
func getCommandLine() string {
	if os.Getenv("GSCTL_DISABLE_CMDLINE_TRACKING") != "" {
		return ""
	}
	args := redactPasswordArgs(os.Args)
	return strings.Join(args, " ")
}

// redactPasswordArgs replaces password in an arguments slice
// with "REDACTED"
func redactPasswordArgs(args []string) []string {
	for index, arg := range args {
		if strings.HasPrefix(arg, "--password=") {
			args[index] = "--password=REDACTED"
		} else if arg == "--password" {
			args[index+1] = "REDACTED"
		} else if len(args) > 1 && args[1] == "login" {
			// this will explicitly only apply to the login command
			if strings.HasPrefix(arg, "-p=") {
				args[index] = "-p=REDACTED"
			} else if arg == "-p" {
				args[index+1] = "REDACTED"
			}
		}
	}
	return args
}

// CreateAuthToken creates an auth token using the latest client
func (w *WrapperV2) CreateAuthToken(email, password string) (*models.V4CreateAuthTokenResponse, error) {
	params := auth_tokens.NewCreateAuthTokenParams().WithBody(&models.V4CreateAuthTokenRequest{
		Email:          email,
		PasswordBase64: base64.StdEncoding.EncodeToString([]byte(password)),
	})

	response, err := w.gsclient.AuthTokens.CreateAuthToken(params, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}

// DeleteAuthToken calls the deleteAuthToken operation in the latest client
func (w *WrapperV2) DeleteAuthToken(authToken string) (*models.V4GenericResponse, error) {
	params := auth_tokens.NewDeleteAuthTokenParams().WithAuthorization("giantswarm " + authToken)

	response, err := w.gsclient.AuthTokens.DeleteAuthToken(params, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return response.Payload, nil
}
