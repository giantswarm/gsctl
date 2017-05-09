package client

import (
	"time"

	"github.com/giantswarm/gsclientgen"

	"github.com/giantswarm/gsctl/config"
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

// NewClient allows to create a new API client
// with specific configuration
func NewClient(clientConfig Configuration) *gsclientgen.DefaultApi {
	configuration := gsclientgen.NewConfiguration()
	configuration.BasePath = clientConfig.Endpoint
	configuration.UserAgent = clientConfig.UserAgent
	configuration.Timeout = &config.DefaultTimeout
	if clientConfig.Timeout != 0 {
		configuration.Timeout = &clientConfig.Timeout
	}

	return &gsclientgen.DefaultApi{
		Configuration: configuration,
	}
}
