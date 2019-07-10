package oidc

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"
	"github.com/gobuffalo/packr"
	"github.com/skratchdot/open-golang/open"
)

var (
	templates = packr.NewBox("html")
)

const (
	scope                = "openid email profile user_metadata https://giantswarm.io offline_access"
	clientID             = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	tokenURL             = "https://giantswarm.eu.auth0.com/oauth/token"
	redirectURL          = "http://localhost:8085/oauth/callback"
	authorizationURLBase = "https://giantswarm.eu.auth0.com/authorize"
)

// PKCEResponse represents the result we get from the PKCE flow.
type PKCEResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        string `json:"expires_in"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RunPKCE starts the Authorization Code Grant Flow with PKCE.
// It does roughly the following steps:
// 1. Craft the authorization URL and open the users browser.
// 2. Starting a callback server to wait for the redirect with the code.
// 3. Exchanging the code for an access token and id token.
func RunPKCE(audience string) (PKCEResponse, error) {
	// Construct the authorization url.
	//    1. Generate and store a random codeVerifier.
	codeVerifier := base64URLEncode(fmt.Sprint(rand.Int31()))

	//    2. Using the codeVerifier, generate a sha256 hashed codeChallenge that
	//       will be sent in the authorization request.
	codeChallenge := sha256.Sum256([]byte(codeVerifier))

	//    3. Use the codeChallenge to make a authorizationURL
	authorizationURL := authorizationURL(string(codeChallenge[:]), audience)

	// Open the authorization url in the user's browser, which will eventually
	// redirect the user to the local webserver we'll create next.
	open.Run(authorizationURL)

	fmt.Println(color.YellowString("\nYour browser should now be opening:"))
	fmt.Println(authorizationURL + "\n")

	// Start a local webserver that auth0 can redirect to with the code
	// that we will then exchange for the actual id token.
	// Credit: https://medium.com/@int128/shutdown-http-server-by-endpoint-in-go-2a0e2d7f9b8c
	p, err := startCallbackServer("8085", "/oauth/callback", func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		// Get the code that Auth0 gave us.
		code := r.URL.Query().Get("code")
		errorCode := r.URL.Query().Get("error")
		errorDescription := r.URL.Query().Get("error_description")

		if errorCode != "" {
			pkceResponse := PKCEResponse{
				Error:            errorCode,
				ErrorDescription: errorDescription,
			}

			return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.ErrorDescription)
		}

		// We now have the 'code' which we can then finally exchange
		// for a real id token by doing a final request to Auth0 and passing the code
		// along with the codeVerifier we made at the start.
		pkceResponse, err := getToken(code, codeVerifier)
		if err != nil {
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(templates.Bytes("sso_failed.html")))
			return pkceResponse, err
		}

		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(templates.Bytes("sso_complete.html")))
		return pkceResponse, nil
	})
	if err != nil {
		return p.(PKCEResponse), err
	}

	return p.(PKCEResponse), nil
}

// base64URLEncode encodes a string into URL safe base64.
func base64URLEncode(input string) string {
	r := base64.URLEncoding.EncodeToString([]byte(input))
	r = strings.Replace(r, "/", "_", -1)
	r = strings.Replace(r, "+", "-", -1)
	r = strings.Replace(r, "=", "", -1)

	return r
}

// authorizationURL crafts the URL that we need to visit at Auth0.
// It takes the hashed codeChallenge that starts off the authorization code grant
// flow.
func authorizationURL(codeChallenge string, audience string) string {
	r := authorizationURLBase

	params := url.Values{}
	params.Set("audience", audience)
	params.Set("scope", scope)
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("code_challenge", base64URLEncode(codeChallenge))
	params.Set("code_challenge_method", "S256")
	params.Set("redirect_uri", redirectURL)
	r += "?"
	r += params.Encode()

	return r
}

// getToken performs a POST call to auth0 as the final step of the
// Authorization Code Grant Flow with PKCE.
func getToken(code, codeVerifier string) (pkceResponse PKCEResponse, err error) {
	payload := strings.NewReader(fmt.Sprintf(`{
    "grant_type":"authorization_code",
    "client_id": "%s",
    "code_verifier": "%s",
    "code": "%s",
    "redirect_uri": "%s"
  }`, clientID, codeVerifier, code, redirectURL))

	req, err := http.NewRequest("POST", tokenURL, payload)
	if err != nil {
		pkceResponse.Error = "unknown error"
		pkceResponse.ErrorDescription = "Unable to construct POST request for Auth0."
		return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.Error)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		pkceResponse.Error = "unknown error"
		pkceResponse.ErrorDescription = "Unable to perform POST request to Auth0."
		return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.Error)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		pkceResponse.Error = "unknown error"
		pkceResponse.ErrorDescription = "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later."
		return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.Error)
	}

	json.Unmarshal(body, &pkceResponse)

	// This is a real error from Auth0, in this case we have Error and ErrorDescription
	// set by what Auth0 sent us.
	if pkceResponse.Error != "" {
		return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.Error)
	}

	if res.StatusCode != 200 {
		pkceResponse.Error = "unknown error"
		pkceResponse.ErrorDescription = "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later."
		return pkceResponse, microerror.Maskf(authorizationError, pkceResponse.Error)
	}

	return pkceResponse, nil
}
