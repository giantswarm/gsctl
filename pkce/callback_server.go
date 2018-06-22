package pkce

import (
	"log"
	"net/http"
)

// startCallbackServer starts the callback server listening at a specific port
// and path.
//
// It is used as part of the authorization code grant flow, it expects a redirect from
// the authorization server (Auth0) with a code in the query like:
// /?code=XXXXXXXX
//
// Once it recieves this code it calls the provided callback with it, letting
// the callback function handle what to do next (which should be attempting
// to change the code for an access token and id token).
func startCallbackServer(port string, redirectURI string, callback func(code string, w http.ResponseWriter, r *http.Request)) http.Server {
	m := http.NewServeMux()
	s := http.Server{Addr: ":" + port, Handler: m}
	m.HandleFunc(redirectURI, func(w http.ResponseWriter, r *http.Request) {
		// Get the code that Auth0 gave us.
		code := r.URL.Query().Get("code")

		// Send it to the callback function.
		callback(code, w, r)
	})

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return s
}
