package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/gscliauth/config"
	gsclient "github.com/giantswarm/gsclientgen/client"
	"github.com/giantswarm/gsclientgen/client/apps"
	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/client/clusters"
	"github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/gsclientgen/client/key_pairs"
	"github.com/giantswarm/gsclientgen/client/nodepools"
	"github.com/giantswarm/gsclientgen/client/organizations"
	"github.com/giantswarm/gsclientgen/client/releases"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	rootcerts "github.com/hashicorp/go-rootcerts"

	"github.com/giantswarm/gsctl/client/clienterror"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified.
	DefaultTimeout = 10 * time.Second

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Configuration is the client configuration
type Configuration struct {
	// AuthHeaderGetter is a function that returns the header we should use to make API calls.
	AuthHeaderGetter func() (string, error)

	// Endpoint is the base URL of the API.
	Endpoint string

	// Timeout is the maximum time to wait for API requests to succeed.
	Timeout time.Duration

	// UserAgent identifier
	UserAgent string

	// ActivityName identifies the user action through the according header.
	ActivityName string
}

// Wrapper is the structure holding representing our latest API client.
type Wrapper struct {
	// conf is the Configuration used when creating this.
	conf *Configuration

	// gsclient is a pointer to the API client library's client.
	gsclient *gsclient.Gsclientgen

	// requestID is the default request ID to use, can be overridden per request.
	requestID string

	// commandLine is the command line use to execute gsctl, can be overridden.
	commandLine string

	// rawClient is a client used to make simple, authenticated HTTP requests
	// against the Giant Swarm API using Go's net/http API.
	rawClient *http.Client
}

// ClusterStatus is a type we use to unmarshal a cluster status JSON response
// from the API. Note: this is scarce, leaving out many available details.
type ClusterStatus struct {
	Cluster *v1alpha1.StatusCluster `json:"cluster,omitempty"`
}

// New creates a client based on the latest gsclientgen version.
func New(conf *Configuration) (*Wrapper, error) {
	if conf.AuthHeaderGetter == nil {
		conf.AuthHeaderGetter = func() (string, error) { return "", nil }
	}
	if conf.Endpoint == "" {
		return nil, microerror.Mask(endpointNotSpecifiedError)
	}

	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, microerror.Mask(endpointInvalidError)
	}

	tlsConfig := &tls.Config{}
	err = rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}
	transport.Transport = setUserAgent(transport.Transport, conf.UserAgent)

	rawClient := &http.Client{
		Transport: transport.Transport,
		Timeout:   conf.Timeout,
	}

	return &Wrapper{
		conf:        conf,
		gsclient:    gsclient.New(transport, strfmt.Default),
		requestID:   randomRequestID(),
		commandLine: getCommandLine(),
		rawClient:   rawClient,
	}, nil
}

// NewWithConfig creates a new client wrapper for a certain endpoint, optionally
// using a certain auth token.
func NewWithConfig(endpointString, token string) (*Wrapper, error) {
	endpoint := config.Config.ChooseEndpoint(endpointString)
	ClientConfig := &Configuration{
		AuthHeaderGetter: config.Config.AuthHeaderGetter(endpoint, token),
		Endpoint:         endpoint,
		Timeout:          20 * time.Second,
		UserAgent:        config.UserAgent(),
	}

	return New(ClientConfig)
}

type roundTripperWithUserAgent struct {
	inner http.RoundTripper
	Agent string
}

// RoundTrip overwrites the http.RoundTripper.RoundTrip function to add our
// User-Agent HTTP header to a request.
func (rtwua *roundTripperWithUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", rtwua.Agent)
	return rtwua.inner.RoundTrip(r)
}

// setUserAgent sets the User-Agent header value for subsequent requests
// made using this transport.
func setUserAgent(inner http.RoundTripper, userAgent string) http.RoundTripper {
	return &roundTripperWithUserAgent{
		inner: inner,
		Agent: userAgent,
	}
}

// randomRequestID returns a new request ID.
func randomRequestID() string {
	size := 14
	b := make([]rune, size)
	for i := range b {
		b[i] = randomStringCharset[rand.Intn(len(randomStringCharset))]
	}
	return string(b)
}

// getCommandLine returns the command line that has been called.
func getCommandLine() string {
	if os.Getenv("GSCTL_DISABLE_CMDLINE_TRACKING") != "" {
		return ""
	}
	args := redactPasswordArgs(os.Args)
	return strings.Join(args, " ")
}

// redactPasswordArgs replaces password in an arguments slice
// with "REDACTED".
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

// AuxiliaryParams are parameters that can be passed to API calls optionally.
type AuxiliaryParams struct {
	CommandLine  string
	RequestID    string
	ActivityName string
	Timeout      time.Duration
}

// DefaultAuxiliaryParams returns a partially pre-populated AuxiliaryParams
// object.
func (w *Wrapper) DefaultAuxiliaryParams() *AuxiliaryParams {
	return &AuxiliaryParams{
		CommandLine: getCommandLine(),
		RequestID:   randomRequestID(),
	}
}

// GetConfiguration returns the client wrapper's configuration, e. g. for debugging purposes.
func (w *Wrapper) GetConfiguration() *Configuration {
	return w.conf
}

// paramSetter is the interface we use to abstract away the differences between
// request parameter types.
type paramSetter interface {
	SetTimeout(time.Duration)
	SetXGiantSwarmActivity(*string)
	SetXRequestID(*string)
	SetXGiantSwarmCmdLine(*string)
}

// setParams takes parameters from an AuxiliaryParams input, and from the
// client wrapper (or rather it's config) and sets request parameters
// accordingly, independent of type.
func setParams(p *AuxiliaryParams, w *Wrapper, params paramSetter) {
	// first take client-level config params
	if w != nil && w.conf != nil {
		if w.conf.Timeout > 0 {
			params.SetTimeout(w.conf.Timeout)
		}
		if w.commandLine != "" {
			params.SetXGiantSwarmCmdLine(&w.commandLine)
		}
		if w.conf.ActivityName != "" {
			params.SetXGiantSwarmActivity(&w.conf.ActivityName)
		}
		if w.requestID != "" {
			params.SetXRequestID(&w.requestID)
		}
	}

	// let per-request params overwrite the above
	if p != nil {
		if p.Timeout > 0 {
			params.SetTimeout(p.Timeout)
		}
		if p.CommandLine != "" {
			params.SetXGiantSwarmCmdLine(&p.CommandLine)
		}
		if p.ActivityName != "" {
			params.SetXGiantSwarmActivity(&p.ActivityName)
		}
		if p.RequestID != "" {
			params.SetXRequestID(&p.RequestID)
		}
	}
}

func getAuthorization(w *Wrapper) (runtime.ClientAuthInfoWriter, error) {
	authHeader, err := w.conf.AuthHeaderGetter()
	if err != nil {
		return nil, err
	}

	return httptransport.APIKeyAuth("Authorization", "header", authHeader), nil
}

// CreateAuthToken creates an auth token using the gsclientgen client.
func (w *Wrapper) CreateAuthToken(email, password string, p *AuxiliaryParams) (*auth_tokens.CreateAuthTokenOK, error) {
	params := auth_tokens.NewCreateAuthTokenParams().WithBody(&models.V4CreateAuthTokenRequest{
		Email:          email,
		PasswordBase64: base64.StdEncoding.EncodeToString([]byte(password)),
	})
	setParams(p, w, params)

	response, err := w.gsclient.AuthTokens.CreateAuthToken(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// DeleteAuthToken calls the API's deleteAuthToken operation using the gsclientgen client.
func (w *Wrapper) DeleteAuthToken(authToken string, p *AuxiliaryParams) (*auth_tokens.DeleteAuthTokenOK, error) {

	params := auth_tokens.NewDeleteAuthTokenParams()
	setParams(p, w, params)

	response, err := w.gsclient.AuthTokens.DeleteAuthToken(params, httptransport.APIKeyAuth("Authorization", "header", "giantswarm "+authToken))
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// CreateCluster creates cluster using the gsclientgen client.
func (w *Wrapper) CreateCluster(addClusterRequest *models.V4AddClusterRequest, p *AuxiliaryParams) (*clusters.AddClusterCreated, error) {
	params := clusters.NewAddClusterParams().WithBody(addClusterRequest)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.AddCluster(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// ModifyCluster modifies a cluster using the gsclientgen client.
func (w *Wrapper) ModifyCluster(clusterID string, body *models.V4ModifyClusterRequest, p *AuxiliaryParams) (*clusters.ModifyClusterOK, error) {
	params := clusters.NewModifyClusterParams().WithClusterID(clusterID).WithBody(body)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.ModifyCluster(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// DeleteCluster deletes a cluster using the gsclientgen client.
func (w *Wrapper) DeleteCluster(clusterID string, p *AuxiliaryParams) (*clusters.DeleteClusterAccepted, error) {
	params := clusters.NewDeleteClusterParams().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.DeleteCluster(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetClusters fetches a list of clusters using the gsclientgen client.
func (w *Wrapper) GetClusters(p *AuxiliaryParams) (*clusters.GetClustersOK, error) {
	params := clusters.NewGetClustersParams()
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusters(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetClusterV4 fetches details on a V4 cluster.
func (w *Wrapper) GetClusterV4(clusterID string, p *AuxiliaryParams) (*clusters.GetClusterOK, error) {
	params := clusters.NewGetClusterParams().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetCluster(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetClusterV5 fetches details on a V5 cluster.
func (w *Wrapper) GetClusterV5(clusterID string, p *AuxiliaryParams) (*clusters.GetClusterV5OK, error) {
	params := clusters.NewGetClusterV5Params().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusterV5(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// CreateNodePool creates a node pool.
func (w *Wrapper) CreateNodePool(clusterID string, addNodePoolRequest *models.V5AddNodePoolRequest, p *AuxiliaryParams) (*nodepools.AddNodePoolCreated, error) {
	params := nodepools.NewAddNodePoolParams().WithBody(addNodePoolRequest).WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Nodepools.AddNodePool(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetNodePool fetches a node pool.
func (w *Wrapper) GetNodePool(clusterID, nodePoolID string, p *AuxiliaryParams) (*nodepools.GetNodePoolOK, error) {
	params := nodepools.NewGetNodePoolParams().WithClusterID(clusterID).WithNodepoolID(nodePoolID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Nodepools.GetNodePool(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetNodePools fetches a list of node pools.
func (w *Wrapper) GetNodePools(clusterID string, p *AuxiliaryParams) (*nodepools.GetNodePoolsOK, error) {
	params := nodepools.NewGetNodePoolsParams().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Nodepools.GetNodePools(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// ModifyNodePool modifies a node pool.
func (w *Wrapper) ModifyNodePool(clusterID, nodePoolID string, modifyNodePoolRequest *models.V5ModifyNodePoolRequest, p *AuxiliaryParams) (*nodepools.ModifyNodePoolOK, error) {
	params := nodepools.NewModifyNodePoolParams().WithClusterID(clusterID).WithNodepoolID(nodePoolID).WithBody(modifyNodePoolRequest)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Nodepools.ModifyNodePool(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// DeleteNodePool deletes a node pool.
func (w *Wrapper) DeleteNodePool(clusterID, nodePoolID string, p *AuxiliaryParams) (*nodepools.DeleteNodePoolAccepted, error) {
	params := nodepools.NewDeleteNodePoolParams().WithClusterID(clusterID).WithNodepoolID(nodePoolID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Nodepools.DeleteNodePool(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetDefaultCluster determines which is the default cluster.
// If only one cluster exists, the default cluster is the only cluster.
func (w *Wrapper) GetDefaultCluster(p *AuxiliaryParams) (string, error) {
	params := clusters.NewGetClustersParams()
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return "", microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusters(params, authWriter)
	if err != nil {
		return "", clienterror.New(err)
	}

	if len(response.Payload) == 1 {
		return response.Payload[0].ID, nil
	}

	return "", nil
}

// CreateKeyPair calls the addKeyPair API operation using the gsclientgen client.
func (w *Wrapper) CreateKeyPair(clusterID string, addKeyPairRequest *models.V4AddKeyPairRequest, p *AuxiliaryParams) (*key_pairs.AddKeyPairOK, error) {
	params := key_pairs.NewAddKeyPairParams().WithClusterID(clusterID).WithBody(addKeyPairRequest)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.KeyPairs.AddKeyPair(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetKeyPairs calls the API to fetch key pairs using the gsclientgen client.
func (w *Wrapper) GetKeyPairs(clusterID string, p *AuxiliaryParams) (*key_pairs.GetKeyPairsOK, error) {
	params := key_pairs.NewGetKeyPairsParams().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.KeyPairs.GetKeyPairs(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetInfo calls the API's getInfo operation using the gsclientgen client.
func (w *Wrapper) GetInfo(p *AuxiliaryParams) (*info.GetInfoOK, error) {
	params := info.NewGetInfoParams()
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Info.GetInfo(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetReleases calls the API's getReleases operation using the gsclientgen client.
func (w *Wrapper) GetReleases(p *AuxiliaryParams) (*releases.GetReleasesOK, error) {
	params := releases.NewGetReleasesParams()
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Releases.GetReleases(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetOrganizations calls the API's getOrganizations operation using the gsclientgen client.
func (w *Wrapper) GetOrganizations(p *AuxiliaryParams) (*organizations.GetOrganizationsOK, error) {
	params := organizations.NewGetOrganizationsParams()
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.GetOrganizations(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetCredential calls the API's getCredential operation using the gsclientgen client.
func (w *Wrapper) GetCredential(organizationID string, credentialID string, p *AuxiliaryParams) (*organizations.GetCredentialOK, error) {
	params := organizations.NewGetCredentialParams().WithOrganizationID(organizationID).WithCredentialID(credentialID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.GetCredential(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// SetCredentials calls the API's addCredentials operation of an organization.
func (w *Wrapper) SetCredentials(organizationID string, addCredentialsRequest *models.V4AddCredentialsRequest, p *AuxiliaryParams) (*organizations.AddCredentialsCreated, error) {
	params := organizations.NewAddCredentialsParams().WithOrganizationID(organizationID).WithBody(addCredentialsRequest)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.AddCredentials(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetClusterStatus fetches details on a cluster using the gsclientgen client.
func (w *Wrapper) GetClusterStatus(clusterID string, p *AuxiliaryParams) (*ClusterStatus, error) {
	params := clusters.NewGetClusterStatusParams().WithClusterID(clusterID)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusterStatus(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	m, err := json.Marshal(response.Payload)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var status ClusterStatus

	err = json.Unmarshal(m, &status)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &status, nil
}

// CreateApp creates an App in a cluster.
func (w *Wrapper) CreateApp(clusterID string, appName string, addAppRequest *models.V4CreateAppRequest, p *AuxiliaryParams) (*apps.CreateClusterAppOK, error) {

	params := apps.NewCreateClusterAppParams().WithClusterID(clusterID).WithAppName(appName).WithBody(addAppRequest)

	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Apps.CreateClusterApp(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

//GetApp fetches details on a cluster using the gsclientgen client.
func (w *Wrapper) GetApp(clusterID string, appName string, p *AuxiliaryParams) (*models.V4GetClusterAppsResponseItems, error) {

	params := apps.NewGetClusterAppsParams().WithClusterID(clusterID)

	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Apps.GetClusterApps(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	apps := response.Payload

	for _, app := range apps {
		if app.Metadata.Name == appName {
			return app, nil
		}
	}

	return nil, nil
}

// GetAppStatus fetches details on a cluster using the gsclientgen client.
func (w *Wrapper) GetAppStatus(clusterID string, appName string, p *AuxiliaryParams) (string, error) {

	params := apps.NewGetClusterAppsParams().WithClusterID(clusterID)

	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return "", microerror.Mask(err)
	}

	response, err := w.gsclient.Apps.GetClusterApps(params, authWriter)
	if err != nil {
		return "", clienterror.New(err)
	}

	//type V4GetClusterAppsResponse []*V4GetClusterAppsResponseItems
	apps := response.Payload

	for _, app := range apps {
		if app.Metadata.Name == appName {
			return app.Status.Release.Status, nil
		}
	}

	return "", nil
}

// DeleteApp deletes an app using the gsclientgen client.
func (w *Wrapper) DeleteApp(clusterID string, appName string, p *AuxiliaryParams) (*apps.DeleteClusterAppOK, error) {
	params := apps.NewDeleteClusterAppParams().WithClusterID(clusterID).WithAppName(appName)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Apps.DeleteClusterApp(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// ModifyApp modifies an app using the gsclientgen client.
func (w *Wrapper) ModifyApp(clusterID string, appName string, body *models.V4ModifyAppRequest, p *AuxiliaryParams) (*apps.ModifyClusterAppOK, error) {

	params := apps.NewModifyClusterAppParams().WithClusterID(clusterID).WithAppName(appName).WithBody(body)
	setParams(p, w, params)

	authWriter, err := getAuthorization(w)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Apps.ModifyClusterApp(params, authWriter)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}
