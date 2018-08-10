package oidc

import (
	"context"
	"net/http"
)

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
	var r CallbackResult
	select {
	case r = <-resultCh:
	}

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
