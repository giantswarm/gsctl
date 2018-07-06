package client

import (
	"crypto/tls"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	rootcerts "github.com/hashicorp/go-rootcerts"
	gsclientgenv1 "gopkg.in/giantswarm/gsclientgen.v1"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified
	DefaultTimeout time.Duration = 60 * time.Second

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

// Configuration is the configuration to be used when creating a new API client
type Configuration struct {
	// AuthHeader is the header we should use to make API calls
	AuthHeader string
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

type ClientWrapper struct {
	authHeader string
	client     *gsclientgenv1.DefaultApi
	requestID  string
}

// NewClient allows to create a new API client
// with specific configuration
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

func init() {
	rand.Seed(time.Now().UnixNano())
	requestIDHeader = randomRequestID()
	cmdLine = getCommandLine()
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

func (c *ClientWrapper) AddCluster(body gsclientgenv1.V4AddClusterRequest, activityName string) (*gsclientgenv1.V4GenericResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.AddCluster(c.authHeader, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

func (c *ClientWrapper) AddKeyPair(clusterID string, body gsclientgenv1.V4AddKeyPairBody, activityName string) (*gsclientgenv1.V4AddKeyPairResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.AddKeyPair(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

func (c *ClientWrapper) DeleteCluster(clusterID string, activityName string) (*gsclientgenv1.V4GenericResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.DeleteCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}

func (c *ClientWrapper) GetCluster(clusterID string, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetCluster(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) GetClusters(activityName string) ([]gsclientgenv1.V4ClusterListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetClusters(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) GetInfo(activityName string) (*gsclientgenv1.V4InfoResponse, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetInfo(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) GetKeyPairs(clusterID string, activityName string) ([]gsclientgenv1.KeyPairModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetKeyPairs(c.authHeader, clusterID, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) GetReleases(activityName string) ([]gsclientgenv1.V4ReleaseListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetReleases(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) GetUserOrganizations(activityName string) ([]gsclientgenv1.V4OrganizationListItem, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.GetUserOrganizations(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) ModifyCluster(clusterID string, body gsclientgenv1.V4ModifyClusterRequest, activityName string) (*gsclientgenv1.V4ClusterDetailsModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.ModifyCluster(c.authHeader, clusterID, body, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
func (c *ClientWrapper) UserLogin(email string, body gsclientgenv1.LoginBodyModel, activityname string) (*gsclientgenv1.LoginResponseModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.UserLogin(email, body, c.requestID, activityname, cmdLine)
	return response, apiResponse, err
}
func (c *ClientWrapper) UserLogout(activityName string) (*gsclientgenv1.GenericResponseModel, *gsclientgenv1.APIResponse, error) {
	response, apiResponse, err := c.client.UserLogout(c.authHeader, c.requestID, activityName, getCommandLine())
	return response, apiResponse, err
}
