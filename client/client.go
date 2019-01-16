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

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	rootcerts "github.com/hashicorp/go-rootcerts"

	gsclient "github.com/giantswarm/gsclientgen/client"
	"github.com/giantswarm/gsclientgen/client/auth_tokens"
	"github.com/giantswarm/gsclientgen/client/clusters"
	"github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/gsclientgen/client/key_pairs"
	"github.com/giantswarm/gsclientgen/client/organizations"
	"github.com/giantswarm/gsclientgen/client/releases"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client/clienterror"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified.
	DefaultTimeout = 60 * time.Second

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

// WrapperV2 is the structure holding representing our latest API client.
type WrapperV2 struct {
	// conf is the Configuration used when creating this.
	conf *Configuration

	// gsclient is a pointer to the API client library's client.
	gsclient *gsclient.Gsclientgen

	// requestID is the default request ID to use, can be overridden per request.
	requestID string

	// commandLine is the command line use to execute gsctl, can be overridden.
	commandLine string
}

// NewV2 creates a client based on the latest gsclientgen version.
func NewV2(conf *Configuration) (*WrapperV2, error) {
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

	return &WrapperV2{
		conf:        conf,
		gsclient:    gsclient.New(transport, strfmt.Default),
		requestID:   randomRequestID(),
		commandLine: getCommandLine(),
	}, nil
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
func (w *WrapperV2) DefaultAuxiliaryParams() *AuxiliaryParams {
	return &AuxiliaryParams{
		CommandLine: getCommandLine(),
		RequestID:   randomRequestID(),
	}
}

// paramSetter is the interface we use to abstract away the differences between
// request parameter types.
type paramSetter interface {
	SetTimeout(time.Duration)
	SetXGiantSwarmActivity(*string)
	SetXRequestID(*string)
	SetXGiantSwarmCmdLine(*string)
}

// paramSetterWithAuthorization is the interface we use to abstract
// parameter types that include a SetAuthorization function.
type paramSetterWithAuthorization interface {
	paramSetter

	SetAuthorization(string)
}

// setParams takes parameters from an AuxiliaryParams input, and from the
// client wrapper (or rather it's config) and sets request parameters
// accordingly, independent of type.
func setParams(p *AuxiliaryParams, w *WrapperV2, params paramSetter) {
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

// setParamsWithAuthorization does the same as setParams, but also sets
// an Autorization header if configured.
func setParamsWithAuthorization(p *AuxiliaryParams, w *WrapperV2, params paramSetterWithAuthorization) error {
	if w != nil && w.conf != nil {
		authHeader, err := w.conf.AuthHeaderGetter()
		if err != nil {
			return err
		}

		if authHeader != "" {
			params.SetAuthorization(authHeader)
		}
	}

	setParams(p, w, params)

	return nil
}

// CreateAuthToken creates an auth token using the V2 client.
func (w *WrapperV2) CreateAuthToken(email, password string, p *AuxiliaryParams) (*auth_tokens.CreateAuthTokenOK, error) {
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

// DeleteAuthToken calls the API's deleteAuthToken operation using the V2 client.
func (w *WrapperV2) DeleteAuthToken(authToken string, p *AuxiliaryParams) (*auth_tokens.DeleteAuthTokenOK, error) {
	params := auth_tokens.NewDeleteAuthTokenParams().WithAuthorization("giantswarm " + authToken)
	setParams(p, w, params)

	response, err := w.gsclient.AuthTokens.DeleteAuthToken(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// CreateCluster creates cluster using the V2 client.
func (w *WrapperV2) CreateCluster(addClusterRequest *models.V4AddClusterRequest, p *AuxiliaryParams) (*clusters.AddClusterCreated, error) {
	params := clusters.NewAddClusterParams().WithBody(addClusterRequest)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.AddCluster(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// ModifyCluster modifies a cluster using the V2 client.
func (w *WrapperV2) ModifyCluster(clusterID string, body *models.V4ModifyClusterRequest, p *AuxiliaryParams) (*clusters.ModifyClusterOK, error) {
	params := clusters.NewModifyClusterParams().WithClusterID(clusterID).WithBody(body)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.ModifyCluster(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// DeleteCluster deletes a cluster using the V2 client.
func (w *WrapperV2) DeleteCluster(clusterID string, p *AuxiliaryParams) (*clusters.DeleteClusterAccepted, error) {
	params := clusters.NewDeleteClusterParams().WithClusterID(clusterID)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.DeleteCluster(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetClusters fetches details on a cluster using the V2 client.
func (w *WrapperV2) GetClusters(p *AuxiliaryParams) (*clusters.GetClustersOK, error) {
	params := clusters.NewGetClustersParams()
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusters(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetCluster fetches details on a cluster using the V2 client.
func (w *WrapperV2) GetCluster(clusterID string, p *AuxiliaryParams) (*clusters.GetClusterOK, error) {
	params := clusters.NewGetClusterParams().WithClusterID(clusterID)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetCluster(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// CreateKeyPair calls the addKeyPair API operation using the V2 client.
func (w *WrapperV2) CreateKeyPair(clusterID string, addKeyPairRequest *models.V4AddKeyPairRequest, p *AuxiliaryParams) (*key_pairs.AddKeyPairOK, error) {
	params := key_pairs.NewAddKeyPairParams().WithClusterID(clusterID).WithBody(addKeyPairRequest)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.KeyPairs.AddKeyPair(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetKeyPairs calls the API to fetch key pairs using the V2 client.
func (w *WrapperV2) GetKeyPairs(clusterID string, p *AuxiliaryParams) (*key_pairs.GetKeyPairsOK, error) {
	params := key_pairs.NewGetKeyPairsParams().WithClusterID(clusterID)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.KeyPairs.GetKeyPairs(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetInfo calls the API's getInfo operation using the V2 client.
func (w *WrapperV2) GetInfo(p *AuxiliaryParams) (*info.GetInfoOK, error) {
	params := info.NewGetInfoParams()
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Info.GetInfo(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetReleases calls the API's getReleases operation using the V2 client.
func (w *WrapperV2) GetReleases(p *AuxiliaryParams) (*releases.GetReleasesOK, error) {
	params := releases.NewGetReleasesParams()
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Releases.GetReleases(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetOrganizations calls the API's getOrganizations operation using the new client.
func (w *WrapperV2) GetOrganizations(p *AuxiliaryParams) (*organizations.GetOrganizationsOK, error) {
	params := organizations.NewGetOrganizationsParams()
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.GetOrganizations(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetCredential calls the API's getCredential operation using the V2 client.
func (w *WrapperV2) GetCredential(organizationID string, credentialID string, p *AuxiliaryParams) (*organizations.GetCredentialOK, error) {
	params := organizations.NewGetCredentialParams().WithOrganizationID(organizationID).WithCredentialID(credentialID)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.GetCredential(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// SetCredentials calls the API's addCredentials operation of an organization.
func (w *WrapperV2) SetCredentials(organizationID string, addCredentialsRequest *models.V4AddCredentialsRequest, p *AuxiliaryParams) (*organizations.AddCredentialsCreated, error) {
	params := organizations.NewAddCredentialsParams().WithOrganizationID(organizationID).WithBody(addCredentialsRequest)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Organizations.AddCredentials(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}

// GetCluster fetches details on a cluster using the V2 client.
func (w *WrapperV2) GetClusterStatus(clusterID string, p *AuxiliaryParams) (*clusters.GetClusterStatusOK, error) {
	params := clusters.NewGetClusterStatusParams().WithClusterID(clusterID)
	err := setParamsWithAuthorization(p, w, params)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	response, err := w.gsclient.Clusters.GetClusterStatus(params, nil)
	if err != nil {
		return nil, clienterror.New(err)
	}

	return response, nil
}
