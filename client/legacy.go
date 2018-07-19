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

// AddKeyPair calls the addKeyPair operation via the old client.
func (c *Wrapper) AddKeyPair(clusterID string, body gsclientgenv1.V4AddKeyPairBody, activityName string) (*gsclientgenv1.V4AddKeyPairResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.AddKeyPair(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// DeleteCluster calls the deleteCluster operation via the old client.
func (c *Wrapper) DeleteCluster(clusterID string, activityName string) (*gsclientgenv1.V4GenericResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.DeleteCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetCluster calls the getCluster operation via the old client.
func (c *Wrapper) GetCluster(clusterID string, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetClusters calls the getClusters operation via the old client.
func (c *Wrapper) GetClusters(activityName string) ([]gsclientgenv1.V4ClusterListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetClusters(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetInfo calls the getInfo operation via the old client.
func (c *Wrapper) GetInfo(activityName string) (*gsclientgenv1.V4InfoResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetInfo(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetKeyPairs calls the getKeyPairs operation via the old client.
func (c *Wrapper) GetKeyPairs(clusterID string, activityName string) ([]gsclientgenv1.KeyPairModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetKeyPairs(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetReleases calls the getReleases operation via the old client.
func (c *Wrapper) GetReleases(activityName string) ([]gsclientgenv1.V4ReleaseListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetReleases(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetUserOrganizations calls the getUserOrganization operation via the old client.
func (c *Wrapper) GetUserOrganizations(activityName string) ([]gsclientgenv1.V4OrganizationListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetUserOrganizations(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// ModifyCluster calls the modifyCluster operation via the old client.
func (c *Wrapper) ModifyCluster(clusterID string, body gsclientgenv1.V4ModifyClusterRequest, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.ModifyCluster(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
