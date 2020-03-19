package login

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	// TODO: Inject these at build time
	clientID     = "dex-k8s-authenticator"
	clientSecret = "LS0tLS1CRUdJTiBQR1AgU0lHTkFUVVJFLS0tLS0KCnd4NEVCd01JUUZ6WDZpais1cEZnUXFYNHV3b1RIWkh2R3BPNm5oOTVod1RTNEFIa1JwbmMwazZJQ1Nodzh5RmIKTlk4eG8rRVJ4T0RDNEpMaHB6L2d0T0l3aDNJQzRQUG05RlpxL2ovdEJRNE05Y1BhR3hhVForOUEraXVPMGRmbQpwaHVtUno0a1VzOUpFa3RZMkd2ZGpaVHloRUF2REZqTW55MmN4aHppdzhpM1F5UFF2UFU2aE9DMjVCVGpsRHFkCmtsUVVRYnJXUE9WeFAxdmliZTNjc3VIZ2pBQT0KPUZ5RmMKLS0tLS1FTkQgUEdQIFNJR05BVFVSRS0tLS0t"
)

type Authenticator struct {
	provider     *oidc.Provider
	clientConfig oauth2.Config
	ctx          context.Context
}

func newAuthenticator() (*Authenticator, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://dex.g8s.ginger.eu-west-1.aws.gigantic.io")
	if err != nil {
		log.Fatalf("failed to get provider: %v", err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "https://dex.g8s.ginger.eu-west-1.aws.gigantic.io/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		provider:     provider,
		clientConfig: config,
		ctx:          ctx,
	}, nil
}

type AuthResult struct {
	OAuth2Token   *oauth2.Token
	IDTokenClaims *json.RawMessage
}

func loginOIDC(args Arguments) (loginResult, error) {
	auther, err := newAuthenticator()
	if err != nil {
		log.Fatalf("failed to get authenticator: %v", err)
	}
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, auther.clientConfig.AuthCodeURL("state"), http.StatusFound)
	// })

	p, err := startCallbackServer("443", "/callback", func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		if r.URL.Query().Get("state") != "state" {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return nil, err
		}
		token, err := auther.clientConfig.Exchange(auther.ctx, r.URL.Query().Get("code"))
		if err != nil {
			log.Printf("no token found: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return nil, err
		}

		oidcConfig := &oidc.Config{
			ClientID: clientID,
		}

		idToken, err := auther.provider.Verifier(oidcConfig).Verify(auther.ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		resp := AuthResult{token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}

		return resp, nil
	})
	if err != nil {
		return p.(loginResult), err
	}

	return p.(loginResult), nil

}

// CallbackResult is used by our channel to store callback results.
type CallbackResult struct {
	Interface interface{}
	Error     error
}

// startCallbackServer starts a server listening at a specific path and port and
// calls a callback function as soon as that path is hit.
//
// It blocks and waits until the path is hit, then shutsdown and returns the
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
