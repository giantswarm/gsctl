package login

import (
	"context"
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
	clientID     = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	clientSecret = "WmZq4t7w!z%C*F-JaNdRgUkXp2r5u8x/"
	authURL      = "http://localhost:8085"
	redirectPath = "/oauth/callback"
)

type Authenticator struct {
	provider     *oidc.Provider
	clientConfig oauth2.Config
	ctx          context.Context
}

type AuthResult struct {
	OAuth2Token *oauth2.Token
	IDToken     *oidc.IDToken
	UserInfo    *oidc.UserInfo
}

type Installation struct {
	baseURL  string
	provider string
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

	return i, nil
}

func (installation *Installation) newAuthenticator() (*Authenticator, error) {
	ctx := context.Background()
	issuer := fmt.Sprintf("https://dex.g8s.%s", installation.baseURL)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, microerror.Maskf(err, "Could not access authentication issuer.")
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  authURL + redirectPath,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	return &Authenticator{
		provider:     provider,
		clientConfig: config,
		ctx:          ctx,
	}, nil
}

func (a *Authenticator) handleCallback(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.URL.Query().Get("state") != "state" {
		return nil, microerror.Maskf(authorizationError, "State did not match.")
	}

	token, err := a.clientConfig.Exchange(a.ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "No token found.")
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, microerror.Maskf(authorizationError, "No id_token field in OAuth2 token.")
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	idToken, err := a.provider.Verifier(oidcConfig).Verify(a.ctx, rawIDToken)
	if err != nil {
		return nil, microerror.Maskf(authorizationError, "Failed to verify ID Token.")
	}

	resp := AuthResult{OAuth2Token: token}
	if err := idToken.Claims(&resp.IDToken); err != nil {
		return nil, microerror.Maskf(authorizationError, "Could not construct the ID Token.")
	}
	resp.IDToken = idToken

	userInfo, err := a.provider.UserInfo(a.ctx, a.clientConfig.TokenSource(a.ctx, token))
	if err := userInfo.Claims(&resp.UserInfo); err != nil {
		return nil, microerror.Maskf(authorizationError, "Could not construct the User Info.")
	}
	resp.UserInfo = userInfo

	return resp, nil
}

func loginOIDC(args Arguments) (loginResult, error) {
	result := loginResult{}

	installation, err := newInstallation("ginger.eu-west-1.aws.gigantic.io")
	if err != nil {
		return loginResult{}, microerror.Maskf(err, "Could not access authentication issuer.")
	}
	auther, err := installation.newAuthenticator()
	if err != nil {
		log.Fatalf("failed to get authenticator: %v", err)
	}

	// Open the authorization url in the user's browser, which will eventually
	// redirect the user to the local webserver we'll create next.
	open.Run(authURL)

	p, err := auther.startCallbackServer("8085", redirectPath, auther.handleCallback)
	authResult := p.(AuthResult)

	result = loginResult{
		apiEndpoint:     installation.baseURL,
		email:           authResult.UserInfo.Email,
		loggedOutBefore: false,
		provider:        installation.provider,
		token:           "token",
	}

	return result, nil
}

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
func (a *Authenticator) startCallbackServer(port string, redirectURI string, callback func(w http.ResponseWriter, r *http.Request) (interface{}, error)) (interface{}, error) {
	// Set a channel we will block on and wait for the result.
	resultCh := make(chan CallbackResult)

	// Setup the server.
	m := http.NewServeMux()
	s := &http.Server{Addr: ":" + port, Handler: m}

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, a.clientConfig.AuthCodeURL("state"), http.StatusFound)
	})

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
