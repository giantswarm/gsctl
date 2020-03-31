package login

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os/user"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	oidc "github.com/coreos/go-oidc"
	k "github.com/giantswarm/gsctl/kubernetes"
	"github.com/giantswarm/microerror"
	"github.com/gobuffalo/packr"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const caCert = `-----BEGIN CERTIFICATE-----
MIIDdjCCAl6gAwIBAgIUEm/lmd55cJt+wfbVAr5MmJmg+BswDQYJKoZIhvcNAQEL
BQAwKzEpMCcGA1UEAxMgZ2luZ2VyLmV1LXdlc3QtMS5hd3MuZ2lnYW50aWMuaW8w
HhcNMjAwMTE1MTM1MjQzWhcNMjkxMTIzMTM1MzEzWjArMSkwJwYDVQQDEyBnaW5n
ZXIuZXUtd2VzdC0xLmF3cy5naWdhbnRpYy5pbzCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBALnpflHlKxJ/Hl7J8+5B8inf477sZvfEID4HQfoVC2VZPu4O
P4eoYhQ11yqir5ehmGKgClNYFCEbtWbJwNnOoS8F7/AH+BsRtUNzXYsVj9VCpwvO
hpDpetA4yhfv0sK292HhlIwdFrpeNmaO+sRTz/34aK7RbOfXnJ12VvL/61ppmizj
7f+7MFdRcPdu+yhKThKlLUfnGciHSS2xOS+GJ9wUvtjleZAW6pmZX5sTCafncNJ5
d8AphigZbn3OjNepVelhPnCtNR2kCD6NAxZi+SkGtoZg1EtxMgHhfjrokpolOqTR
jZsnAV3HSGqLYMiDliJLqlzNa9kv2IWvMEwSvMcCAwEAAaOBkTCBjjAOBgNVHQ8B
Af8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU6DIUUBsUv5ae1TQS
rSYQ+oHs0xcwHwYDVR0jBBgwFoAU6DIUUBsUv5ae1TQSrSYQ+oHs0xcwKwYDVR0R
BCQwIoIgZ2luZ2VyLmV1LXdlc3QtMS5hd3MuZ2lnYW50aWMuaW8wDQYJKoZIhvcN
AQELBQADggEBALT1T9v4+5kfDRuFzLoDYX/rZmILVvbItRMAcXV62bsgiK5ko9sh
ro0eBhHmKvmGz70Y4M+dA0mCqlt1m16PYnz96LF4dBvF7/t4by4FzQRpObax9RPl
RmC/xqB285RHOU0gvHM5xeI3KDLapJDh+Al9oH9pfZmLf2Hc/vGjgMdjA1iiNyhn
tpUu65HZSntcmcLR9hlZ6aPMg60dXzoDhKsnTERNLygDq40G3OxQu7Hcejb5Tr/u
mFhWZ+pFznSeD34Jek/irOQ8x8S8LqPZaUCqvGgedkE0APUsg82Elsc2RRf2fGUV
eYyPtbJ0CrKvc2vKFPH+whGPvAkM1z5IXnM=
-----END CERTIFICATE-----`

const (
	clientID     = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	clientSecret = "11dr35cv6v4hkd4gjgzl9f36xj2kyjzdl1i4etclg3s0tgrtx8gb997jfkdn4hg8"

	authCallbackURL  = "http://localhost:8085"
	authCallbackPath = "/oauth/callback"

	apiServerPrefix  = "g8s"
	authServerPrefix = "dex"

	certificateFileName = "k8s-ca.crt"
)

var (
	authScopes = []string{oidc.ScopeOpenID, "profile", "email", "groups", "offline_access"}
	templates  = packr.NewBox("html")
)

type Authenticator struct {
	ctx          context.Context
	provider     *oidc.Provider
	clientConfig oauth2.Config
	state        string
}

type UserInfo struct {
	Email         string
	EmailVerified bool
	IDToken       string
	RefreshToken  string
	IssuerURL     string
	Username      string
	Groups        []string
}

type Installation struct {
	*Authenticator
	FileSystem afero.Fs
	BaseURL    string
	Provider   string
	Alias      string
	CaCert     string
}

func newInstallation(baseURL string, fs afero.Fs) (*Installation, error) {
	caCert := getClusterCert(baseURL)

	i := &Installation{
		BaseURL:    baseURL,
		CaCert:     caCert,
		FileSystem: fs,
	}

	switch {
	case strings.Contains(baseURL, "aws"):
		i.Provider = "aws"

	case strings.Contains(baseURL, "azure"):
		i.Provider = "azure"

	default:
		i.Provider = "kvm"
	}

	urlParts := strings.Split(baseURL, ".")
	if len(urlParts) == 0 {
		return nil, microerror.Maskf(authorizationError, "The installation alias name could not be determined.")
	}
	i.Alias = urlParts[0]

	return i, nil
}

func (i *Installation) newAuthenticator(redirectURL string, authScopes []string) error {
	ctx := context.Background()
	issuer := getUrlFromParts("https://", []string{authServerPrefix, apiServerPrefix, i.BaseURL})
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return microerror.Maskf(err, "Could not access authentication issuer.")
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       authScopes,
	}

	i.Authenticator = &Authenticator{
		provider:     provider,
		clientConfig: config,
		ctx:          ctx,
	}

	return nil
}

func (i *Installation) storeCredentials(u *UserInfo) error {
	err := i.writeCertificate()
	if err != nil {
		return microerror.Mask(err)
	}

	k, err := i.generateKubeConfig(u)
	if err != nil {
		return microerror.Mask(err)
	}

	err = i.writeKubeConfig(k)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (i *Installation) writeCertificate() error {
	certPath, err := getKubeCertPath(i.Alias)
	if err != nil {
		return microerror.Mask(err)
	}

	certFilePath := path.Join(certPath, certificateFileName)

	err = i.FileSystem.MkdirAll(certPath, 0700)
	if err != nil {
		return microerror.Mask(err)
	}

	err = afero.WriteFile(i.FileSystem, certFilePath, []byte(i.CaCert), 0600)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (i *Installation) writeKubeConfig(k *k.KubeConfigValue) error {
	d, err := yaml.Marshal(k)
	if err != nil {
		return microerror.Mask(err)
	}

	kubeConfigPath, err := getKubeConfigPath()
	if err != nil {
		return microerror.Mask(err)
	}

	err = afero.WriteFile(i.FileSystem, kubeConfigPath, d, 0600)
	if err != nil {
		return microerror.Maskf(authorizationError, "Error writing kubeconfig.")
	}

	return nil
}

func (i *Installation) readKubeConfig() (*k.KubeConfigValue, error) {
	kubeConfigPath, err := getKubeConfigPath()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	kFile, err := afero.ReadFile(i.FileSystem, kubeConfigPath)
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "Error reading existing kubeconfig.")
	}

	var kubeConfig k.KubeConfigValue
	{
		err = yaml.Unmarshal(kFile, &kubeConfig)
		if err != nil {
			return nil, microerror.Maskf(authorizationError, "Error reading existing kubeconfig.")
		}
	}

	return &kubeConfig, nil
}

func (i *Installation) generateKubeConfig(u *UserInfo) (*k.KubeConfigValue, error) {
	existing, _ := i.readKubeConfig()
	if existing == nil {
		existing = &k.KubeConfigValue{}
	}

	kUsername := getUsername(i.Alias, u.Username)

	// Set current context
	existing.CurrentContext = kUsername

	certPath, err := getKubeCertPath(i.Alias)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	certPath = path.Join(certPath, certificateFileName)
	// Add cluster to list
	existing.Clusters = appendOrModify(
		existing.Clusters,
		k.KubeconfigNamedCluster{
			Name: i.Alias,
			Cluster: k.KubeconfigCluster{
				Server:               getUrlFromParts("https://", []string{apiServerPrefix, i.BaseURL}),
				CertificateAuthority: certPath,
			},
		},
		"Name",
	).([]k.KubeconfigNamedCluster)

	// Add context to list
	existing.Contexts = appendOrModify(
		existing.Contexts,
		k.KubeconfigNamedContext{
			Name: kUsername,
			Context: k.KubeconfigContext{
				Cluster: i.Alias,
				User:    kUsername,
			},
		},
		"Name",
	).([]k.KubeconfigNamedContext)

	// Add authentication info  to list
	existing.Users = appendOrModify(
		existing.Users,
		k.KubeconfigUser{
			Name: kUsername,
			User: k.KubeconfigUserKeyPair{
				AuthProvider: k.KubeconfigAuthProvider{
					Name: "oidc",
					Config: map[string]string{
						"client-id":      i.Authenticator.clientConfig.ClientID,
						"client-secret":  i.Authenticator.clientConfig.ClientSecret,
						"id-token":       u.IDToken,
						"idp-issuer-url": u.IssuerURL,
						"refresh-token":  u.RefreshToken,
					},
				},
			},
		},
		"Name",
	).([]k.KubeconfigUser)

	return existing, nil
}

func (a *Authenticator) getAuthURL() string {
	return a.clientConfig.AuthCodeURL(a.state, oauth2.AccessTypeOffline)
}

func (a *Authenticator) handleIssResponse(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.URL.Query().Get("state") != a.state {
		return nil, microerror.Maskf(authorizationError, "State did not match.")
	}

	res := UserInfo{}

	// Convert authorization code into a token
	token, err := a.clientConfig.Exchange(a.ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, microerror.Mask(err)
	}
	res.RefreshToken = token.RefreshToken

	var ok bool
	// Generate the ID Token
	if res.IDToken, ok = token.Extra("id_token").(string); !ok {
		return nil, microerror.Mask(err)
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	// Verify if ID Token is valid
	idToken, err := a.provider.Verifier(oidcConfig).Verify(a.ctx, res.IDToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	res.IssuerURL = idToken.Issuer

	var claims struct {
		Email    string   `json:"email"`
		Verified bool     `json:"email_verified"`
		Groups   []string `json:"groups"`
	}
	// Get the user's info
	if err = idToken.Claims(&claims); err != nil {
		return nil, microerror.Mask(err)
	}
	res.Email = claims.Email
	res.EmailVerified = claims.Verified
	res.Groups = claims.Groups
	res.Username = strings.Split(res.Email, "@")[0]

	return res, nil
}

func (a *Authenticator) handleCallback(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	res, err := a.handleIssResponse(w, r)
	if err != nil {
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(templates.Bytes("sso_failed.html")))
	}

	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(templates.Bytes("sso_complete.html")))
	return res, nil
}

func loginOIDC(args Arguments) (loginResult, error) {
	result := loginResult{}

	i, err := newInstallation("ginger.eu-west-1.aws.gigantic.io", afero.NewOsFs())
	if err != nil {
		return loginResult{}, microerror.Maskf(err, "Could not define installation.")
	}

	authURL := authCallbackURL + authCallbackPath
	err = i.newAuthenticator(authURL, authScopes)
	if err != nil {
		log.Fatalf("failed to get authenticator: %v", err)
	}

	i.Authenticator.state = generateState()
	aURL := i.Authenticator.getAuthURL()
	// Open the authorization url in the user's browser, which will eventually
	// redirect the user to the local webserver we'll create next.
	open.Run(aURL)

	p, err := startCallbackServer("8085", authCallbackPath, i.Authenticator.handleCallback)
	if err != nil {
		return loginResult{}, microerror.Mask(err)
	}
	authResult, ok := p.(UserInfo)
	if !ok {
		return loginResult{}, microerror.Mask(err)
	}

	// Store kubeconfig
	err = i.storeCredentials(&authResult)
	if err != nil {
		return loginResult{}, microerror.Maskf(err, "Could not store credentials.")
	}

	result = loginResult{
		apiEndpoint:     i.BaseURL,
		email:           authResult.Email,
		loggedOutBefore: false,
		alias:           i.Alias,
		provider:        i.Provider,
		token:           authResult.IDToken,
	}

	return result, nil
}

func getClusterCert(baseURL string) string {
	// TODO: Make HTTP Call to get this from somewhere
	return caCert
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	return state
}

func getUsername(iAlias, oAuth2Username string) string {
	return strings.Join([]string{oAuth2Username, iAlias}, "-")
}

func getUrlFromParts(protocol string, parts []string) string {
	return protocol + strings.Join(parts, ".")
}

func getKubeConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", microerror.Maskf(authorizationError, "Error getting current OS user.")
	}
	kubeConfigPath := filepath.Join(usr.HomeDir, ".kube", "config")

	return kubeConfigPath, nil
}

func getKubeCertPath(alias string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", microerror.Maskf(authorizationError, "Error getting current OS user.")
	}
	kubeCertPath := path.Join(usr.HomeDir, ".kube", "certs", alias)

	return kubeCertPath, nil
}

func appendOrModify(target interface{}, entry interface{}, compareByField string) interface{} {
	// TODO: Add error handling, stricter checking

	// Check if the target is a slice
	if reflect.TypeOf(target).Kind() == reflect.Slice {
		s := reflect.ValueOf(target)

		var (
			t, e   reflect.Value
			update bool
		)

		// Look through all target entries
		for i := 0; i < s.Len(); i++ {
			t = s.Index(i)
			e = reflect.ValueOf(entry)

			// If the current entry `compareByField` field value is the same
			// as the field value of the one provided to the function
			if t.FieldByName(compareByField).Interface() == e.FieldByName(compareByField).Interface() {
				// Replace the value with the new one
				t.Set(e)

				update = true
			}
		}

		if !update {
			// Add value to the end of the slice
			s = reflect.Append(s, reflect.ValueOf(entry))
		}

		return s.Interface()
	}

	return target
}

// From OIDC Package

// CallbackResult is used by our channel to store callback results.
type CallbackResult struct {
	Interface interface{}
	Error     error
}

// startCallbackServer starts a server listening at a specific path and port and
// calls a callback function as soon as that path is hit.
//
// It blocks and waits until the path is hit, then shuts down and returns the
// result of the callback function.
//
// It can be used as part of a authorization code grant flow, i.e. expecting a redirect from
// the authorization server (Auth0) with a code in the query like:
// /?code=XXXXXXXX, use a callback function to handle what to do next
// (which should be attempting to change the code for an access token and id token).
func startCallbackServer(port string, redirectURI string, callback func(w http.ResponseWriter, r *http.Request) (interface{}, error)) (interface{}, error) {
	// Set a channel we will block on and wait for the result.
	resultCh := make(chan CallbackResult)

	// Setup the server.
	m := http.NewServeMux()
	s := &http.Server{Addr: ":" + port, Handler: m}

	// This is the handler for the path we specified, it calls the provided
	// callback as soon as a request arrives and moves the result of the callback
	// on to the resultCh.
	m.HandleFunc(redirectURI, func(w http.ResponseWriter, r *http.Request) {
		// Got a response, call the callback function.
		i, err := callback(w, r)
		resultCh <- CallbackResult{i, err}
	})

	// Start the server
	go startServer(s)

	// Block till the callback gives us a result.
	r := <-resultCh

	// Shutdown the server.
	s.Shutdown(context.Background())

	// Return the result.
	return r.Interface, r.Error
}

func startServer(s *http.Server) {
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

var authorizationError = &microerror.Error{
	Kind: "authorizationError",
}

// IsAuthorizationError asserts authorizationError.
func IsAuthorizationError(err error) bool {
	return microerror.Cause(err) == authorizationError
}
