package login

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"github.com/giantswarm/microerror"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

const (
	clientID         = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	clientSecret     = "WmZq4t7w!z%C*F-JaNdRgUkXp2r5u8x/"
	authCallbackURL  = "http://localhost:8085/oauth/callback"
	authCallbackPath = "/oauth/callback"
)

var (
	authScopes = []string{oidc.ScopeOpenID, "profile", "email", "groups", "offline_access"}
)

type Authenticator struct {
	provider     *oidc.Provider
	clientConfig oauth2.Config
	state        string
	ctx          context.Context
}

type UserInfo struct {
	Email        string
	IDToken      string
	RefreshToken string
	IssuerURL    string
	ClusterCA    string
}

type Installation struct {
	*Authenticator
	baseURL  string
	provider string
	alias    string
}

func newInstallation(baseURL string) (*Installation, error) {
	i := &Installation{baseURL: baseURL}

	switch {
	case strings.Contains(baseURL, "aws"):
		i.provider = "aws"

	case strings.Contains(baseURL, "azure"):
		i.provider = "azure"

	default:
		i.provider = "kvm"
	}

	urlParts := strings.Split(baseURL, ".")
	if len(urlParts) == 0 {
		return nil, microerror.Maskf(authorizationError, "The installation alias name could not be determined.")
	}
	i.alias = urlParts[0]

	return i, nil
}

func (i *Installation) newAuthenticator(redirectURL string, authScopes []string) error {
	ctx := context.Background()
	issuer := fmt.Sprintf("https://dex.g8s.%s", i.baseURL)
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

func (a *Authenticator) getAuthURL() string {
	return a.clientConfig.AuthCodeURL(a.state, oauth2.AccessTypeOffline)
}

func (a *Authenticator) handleCallback(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.URL.Query().Get("state") != a.state {
		return nil, microerror.Maskf(authorizationError, "State did not match.")
	}

	res := UserInfo{}

	// Convert authorization code into a token
	token, err := a.clientConfig.Exchange(a.ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "No token found.")
	}
	res.RefreshToken = token.RefreshToken

	var ok bool
	// Generate the ID Token
	res.IDToken, ok = token.Extra("id_token").(string)
	if !ok {
		return nil, microerror.Maskf(authorizationError, "No id_token field in OAuth2 token.")
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	// Verify if ID Token is valid
	idToken, err := a.provider.Verifier(oidcConfig).Verify(a.ctx, res.IDToken)
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "Failed to verify ID Token.")
	}
	res.IssuerURL = idToken.Issuer

	// Get the user's info
	userInfo, err := a.provider.UserInfo(a.ctx, a.clientConfig.TokenSource(a.ctx, token))
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "Could not construct the User Info.")
	}
	res.Email = userInfo.Email

	return res, nil
}

func loginOIDC(args Arguments) (loginResult, error) {
	result := loginResult{}

	i, err := newInstallation("ginger.eu-west-1.aws.gigantic.io")
	if err != nil {
		return loginResult{}, microerror.Maskf(err, "Could not define installation.")
	}

	err = i.newAuthenticator(authCallbackURL, authScopes)
	if err != nil {
		log.Fatalf("failed to get authenticator: %v", err)
	}

	i.Authenticator.state = generateState()
	aURL := i.Authenticator.getAuthURL()
	// Open the authorization url in the user's browser, which will eventually
	// redirect the user to the local webserver we'll create next.
	open.Run(aURL)

	p, err := startCallbackServer("8085", authCallbackPath, i.Authenticator.handleCallback)
	authResult := p.(UserInfo)

	// Store kubeconfig
	// err = storeCredentials(&authResult, i)
	// if err != nil {
	// 	return loginResult{}, microerror.Maskf(err, "Could not store credentials.")
	// }

	ff, _ := json.MarshalIndent(authResult, "", "\t")
	fmt.Println(string(ff))

	result = loginResult{
		apiEndpoint:     i.baseURL,
		email:           authResult.Email,
		loggedOutBefore: false,
		alias:           i.alias,
		provider:        i.provider,
		token:           authResult.IDToken,
	}

	return result, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	return state
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

// This goes into config package
// func storeCredentials(a *UserInfo, i *Installation) error {
//
// }

//
// func generateKubeConfig(a *UserInfo) clientcmdapi.Config {
//
// }
