package client

import (
	"encoding/json"
	"time"

	"github.com/giantswarm/gsclientgen"
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
func NewClient(clientConfig Configuration) *gsclientgen.DefaultApi {
	configuration := gsclientgen.NewConfiguration()
	configuration.BasePath = clientConfig.Endpoint
	configuration.UserAgent = clientConfig.UserAgent
	configuration.Timeout = &DefaultTimeout
	if clientConfig.Timeout != 0 {
		configuration.Timeout = &clientConfig.Timeout
	}

	return &gsclientgen.DefaultApi{
		Configuration: configuration,
	}
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
