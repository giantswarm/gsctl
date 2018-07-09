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

// ClientWrapper is the structure representing our old client
// based on gsclientgen.v1
type ClientWrapper struct {
	authHeader string
	client     *gsclientgenv1.DefaultApi
	requestID  string
}

// NewClient allows to create an API client
// with specific configuration based on the old gsclientgen.v1
func NewClient(clientConfig Configuration) (*ClientWrapper, error) {
	configuration := gsclientgenv1.NewConfiguration()

	if clientConfig.Endpoint == "" {
		return &ClientWrapper{}, microerror.Mask(endpointNotSpecifiedError)
	}

	configuration.BasePath = clientConfig.Endpoint
	configuration.UserAgent = clientConfig.UserAgent
	configuration.Timeout = &DefaultTimeout
	if clientConfig.Timeout != 0 {
		configuration.Timeout = &clientConfig.Timeout
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
	configuration.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	client := ClientWrapper{
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

// AddCluster calls the addCLuster operation via the old client
func (c *ClientWrapper) AddCluster(body gsclientgenv1.V4AddClusterRequest, activityName string) (*gsclientgenv1.V4GenericResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.AddCluster(c.authHeader, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// AddKeyPair calls the addKeyPair operation via the old client
func (c *ClientWrapper) AddKeyPair(clusterID string, body gsclientgenv1.V4AddKeyPairBody, activityName string) (*gsclientgenv1.V4AddKeyPairResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.AddKeyPair(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// DeleteCluster calls the deleteCluster operation via the old client
func (c *ClientWrapper) DeleteCluster(clusterID string, activityName string) (*gsclientgenv1.V4GenericResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.DeleteCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetCluster calls the getCluster operation via the old client
func (c *ClientWrapper) GetCluster(clusterID string, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetClusters calls the getClusters operation via the old client
func (c *ClientWrapper) GetClusters(activityName string) ([]gsclientgenv1.V4ClusterListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetClusters(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetInfo calls the getInfo operation via the old client
func (c *ClientWrapper) GetInfo(activityName string) (*gsclientgenv1.V4InfoResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetInfo(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetKeyPairs calls the getKeyPairs operation via the old client
func (c *ClientWrapper) GetKeyPairs(clusterID string, activityName string) ([]gsclientgenv1.KeyPairModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetKeyPairs(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetReleases calls the getReleases operation via the old client
func (c *ClientWrapper) GetReleases(activityName string) ([]gsclientgenv1.V4ReleaseListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetReleases(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// GetUserOrganizations calls the getUserOrganization operation via the old client
func (c *ClientWrapper) GetUserOrganizations(activityName string) ([]gsclientgenv1.V4OrganizationListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetUserOrganizations(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// ModifyCluster calls the modifyCluster operation via the old client
func (c *ClientWrapper) ModifyCluster(clusterID string, body gsclientgenv1.V4ModifyClusterRequest, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.ModifyCluster(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

// UserLogin calls the v1 login endpoint via the old client
func (c *ClientWrapper) UserLogin(email string, body gsclientgenv1.LoginBodyModel, activityname string) (*gsclientgenv1.LoginResponseModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.UserLogin(email, body, c.requestID, activityname, cmdLine)
	return response, apiResponse, err
}

// UserLogout calls the v1 logout endpoint via the old client
// func (c *ClientWrapper) UserLogout(activityName string) (*gsclientgenv1.GenericResponseModel, *gsclientgenv1.APIResponse, error) {
// 	response, apiResponse, err := c.client.UserLogout(c.authHeader, c.requestID, activityName, getCommandLine())
// 	return response, apiResponse, err
// }
