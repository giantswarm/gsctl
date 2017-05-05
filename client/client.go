package client

import (
	"time"

	"github.com/giantswarm/gsclientgen"
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
func NewClient(config Configuration) *gsclientgen.DefaultApi {
	configuration := gsclientgen.NewConfiguration()
	configuration.BasePath = config.Endpoint
	configuration.UserAgent = config.UserAgent
	configuration.Timeout = &config.Timeout

	return &gsclientgen.DefaultApi{
		Configuration: configuration,
	}
}
