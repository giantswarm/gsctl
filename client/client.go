package client

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	rootcerts "github.com/hashicorp/go-rootcerts"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified
	DefaultTimeout time.Duration = 60 * time.Second
)

// Configuration is the configuration to be used when creating a new API client
type Configuration struct {
	// Endpoint is the base URL of the API
	Endpoint string
	// Timeout is the maximum time to wait for API requests to succeed
	Timeout time.Duration
	// UserAgent
	UserAgent string
}

// GenericResponse allows to access details of a generic API response (mostly error messages).
type GenericResponse struct {
	Code    string
	Message string
}

// NewClient allows to create a new API client
// with specific configuration
func NewClient(clientConfig Configuration) (*gsclientgen.APIClient, error) {
	configuration := gsclientgen.NewConfiguration()

	if clientConfig.Endpoint == "" {
		return &gsclientgen.APIClient{}, microerror.Mask(endpointNotSpecifiedError)
	}

	// pass our own configuration to the generated client's config object
	configuration.BasePath = clientConfig.Endpoint
	configuration.UserAgent = clientConfig.UserAgent

	timeout := DefaultTimeout
	if clientConfig.Timeout != 0 {
		timeout = clientConfig.Timeout
	}

	// set up client TLS so that custom CAs are accepted.
	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		return nil, microerror.Mask(rootCertsErr)
	}

	configuration.HTTPClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsConfig,
		},
	}

	return gsclientgen.NewAPIClient(configuration), nil
}

// ParseGenericResponse parses the standard code, message response document into
// a struct with the fields Code and Message.
func ParseGenericResponse(jsonBody []byte) (GenericResponse, error) {
	var response = GenericResponse{}
	err := json.Unmarshal(jsonBody, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
