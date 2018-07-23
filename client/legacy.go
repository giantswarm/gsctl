package client

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"

	"github.com/giantswarm/microerror"
	rootcerts "github.com/hashicorp/go-rootcerts"
	gsclientgenv1 "gopkg.in/giantswarm/gsclientgen.v1"
)

// Wrapper is the structure representing our old client
// based on gsclientgen.v1.
type Wrapper struct {
	authHeader string
	client     *gsclientgenv1.DefaultApi
	requestID  string
}

// New allows to create an API client
// with specific configuration based on the old gsclientgen.v1.
func New(clientConfig Configuration) (*Wrapper, error) {
	configuration := gsclientgenv1.NewConfiguration()

	if clientConfig.Endpoint == "" {
		return &Wrapper{}, microerror.Mask(endpointNotSpecifiedError)
	}

	configuration.BasePath = clientConfig.Endpoint
	configuration.UserAgent = clientConfig.UserAgent
	configuration.Timeout = &DefaultTimeout
	if clientConfig.Timeout != 0 {
		configuration.Timeout = &clientConfig.Timeout
	}

	// set up client TLS so that custom CAs are accepted.
	tlsConfig := &tls.Config{}
	err := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}
	configuration.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	client := Wrapper{
		authHeader: clientConfig.AuthHeader,
		client: &gsclientgenv1.DefaultApi{
			Configuration: configuration,
		},
		requestID: randomRequestID(),
	}

	return &client, nil
}

// GenericResponse allows to access details of a generic API response (mostly error messages).
type GenericResponse struct {
	Code    string
	Message string
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
