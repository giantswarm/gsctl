package commands

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/microerror"
	"github.com/gobuffalo/packr"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var (
	// SSOCommand performs the "sso" function
	SSOCommand = &cobra.Command{
		Use:   "sso",
		Short: "Single Sign on for Admins",
		Long:  `Prints a list of all clusters you have access to`,
		Run:   ssoRunOutput,
	}
)

const (
	activityName = "sso"
	clientID     = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	redirectURI  = "http://localhost:8085/oauth/callback"
)

type ssoArguments struct {
	apiEndpoint string
}

func defaultSSOArguments() ssoArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)

	return ssoArguments{
		apiEndpoint: endpoint,
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	RootCommand.AddCommand(SSOCommand)
}

func ssoRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultSSOArguments()

	// Construct the authorization url.
	// 		1. Generate and store a random codeVerifier.
	codeVerifier := base64URLEncode(fmt.Sprint(rand.Int31()))

	// 		2. Using the codeVerifier, generate a sha256 hashed codeChallenge that
	//    	 will be sent in the authorization request.
	codeChallenge := sha256.Sum256([]byte(codeVerifier))

	// 		3. Use the codeChallenge to make a authorizationURL
	authorizationURL := authorizationURL(string(codeChallenge[:]))

	// Start a local webserver that auth0 can redirect to with the code
	// that we will then exchange for the actual id token.
	// Credit: https://medium.com/@int128/shutdown-http-server-by-endpoint-in-go-2a0e2d7f9b8c
	tokenResponseCh := make(chan tokenResponse)
	callbackServer := startCallbackServer(tokenResponseCh, codeVerifier)

	// Open the authorization url in the users browser, which will eventually
	// redirect the user to the local webserver created above.
	open.Run(authorizationURL)

	fmt.Println(color.YellowString("\nYour browser should now be opening:"))
	fmt.Println(authorizationURL)

	// Block until we recieve a token from auth0. (In other words, localhost:8085
	// is hit thanks to auth0's redirect by the users browser with /?code=XXXXXXXX)
	var tokenResponse tokenResponse
	select {
	case tokenResponse = <-tokenResponseCh:
		// Token response recieved, shutdown.
		callbackServer.Shutdown(context.Background())
	}

	if tokenResponse.Error != "" {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println(tokenResponse.Error + ": " + tokenResponse.ErrorDescription)
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Try to parse the ID Token.
	idToken, err := parseIdToken(tokenResponse.IDToken)
	if err != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to parse the ID Token.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Check if the access token works by fetching the installation's name.
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, err := client.NewClient(clientConfig)
	if err != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to verify token by fetching installation details.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Fetch installation name as alias.
	authHeader := "Bearer " + tokenResponse.AccessToken
	infoResponse, _, infoErr := apiClient.GetInfo(authHeader, requestIDHeader, loginActivityName, cmdLine)
	if infoErr != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to verify token by fetching installation details.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	alias := infoResponse.General.InstallationName

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, idToken.Email, "Bearer", tokenResponse.AccessToken); err != nil {
		fmt.Println(color.RedString("\nSomething went while trying to store the token."))
		fmt.Println(err.Error())
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	fmt.Println(color.GreenString("\nYou are logged in as %s at %s.",
		idToken.Email, args.apiEndpoint))
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
func authorizationURL(codeChallenge string) string {
	r := "https://giantswarm.eu.auth0.com/authorize?"

	params := url.Values{}
	params.Set("audience", "giantswarm-api")
	params.Set("scope", "openid email profile user_metadata https://giantswarm.io offline_access")
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("code_challenge", base64URLEncode(codeChallenge))
	params.Set("code_challenge_method", "S256")
	params.Set("redirect_uri", redirectURI)
	r += params.Encode()

	return r
}

// startCallbackServer starts the callback server listening at port 8085.
// As part of the authorization code grant flow, it expects a redirect from
// the authorization server (Auth0) with a code in the query like:
// /?code=XXXXXXXX
//
// Once it recieves this code it try to exhange it for a token
// and will pass the token along to the given channel.
func startCallbackServer(tokenResponseCh chan tokenResponse, codeVerifier string) http.Server {
	box := packr.NewBox("../html")

	m := http.NewServeMux()
	s := http.Server{Addr: ":8085", Handler: m}
	m.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		// Get the code that Auth0 gave us.
		code := r.URL.Query().Get("code")

		// We now have the 'code' which we can then finally exchange
		// for a real id token by doing a final request to Auth0 and passing the code
		// along with the codeVerifier we made at the start.
		tokenResponse, err := getToken(code, codeVerifier)
		if err != nil {
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(box.Bytes("sso_failed.html")))
		} else {
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(box.Bytes("sso_complete.html")))
		}

		tokenResponseCh <- tokenResponse
	})

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return s
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        string `json:"expires_in"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// getToken performs a POST call to auth0 as the final step of the
// Authorization Code Grant Flow with PKCE.
func getToken(code, codeVerifier string) (tokenResponse tokenResponse, err error) {
	tokenURL := "https://giantswarm.eu.auth0.com/oauth/token"
	payload := strings.NewReader(fmt.Sprintf(`{
		"grant_type":"authorization_code",
		"client_id": "%s",
		"code_verifier": "%s",
		"code": "%s",
		"redirect_uri": "%s"
	}`, clientID, codeVerifier, code, redirectURI))

	req, err := http.NewRequest("POST", tokenURL, payload)
	if err != nil {
		tokenResponse.Error = "unknown_error"
		tokenResponse.ErrorDescription = "Unable to construct POST request for Auth0."
		return tokenResponse, microerror.Maskf(unknownError, tokenResponse.Error)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		tokenResponse.Error = "unknown_error"
		tokenResponse.ErrorDescription = "Unable to perform POST request to Auth0."
		return tokenResponse, microerror.Maskf(unknownError, tokenResponse.Error)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		tokenResponse.Error = "unknown_error"
		tokenResponse.ErrorDescription = "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later."
		return tokenResponse, microerror.Maskf(unknownError, tokenResponse.Error)
	}

	json.Unmarshal(body, &tokenResponse)

	// This is a real error from Auth0, in this case we have Error and ErrorDescription
	// set by what Auth0 sent us.
	if tokenResponse.Error != "" {
		return tokenResponse, microerror.Maskf(unknownError, tokenResponse.Error)
	}

	if res.StatusCode != 200 {
		tokenResponse.Error = "unknown_error"
		tokenResponse.ErrorDescription = "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later."
		return tokenResponse, microerror.Maskf(unknownError, tokenResponse.Error)
	}

	return tokenResponse, nil
}

type IDToken struct {
	Email string
}

func parseIdToken(tokenString string) (token IDToken, err error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	t, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		cert, err := getPemCert(token, "https://giantswarm.eu.auth0.com/.well-known/jwks.json")
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))

		return result, nil
	})

	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		if claims["email"] != nil {
			token.Email = claims["email"].(string)
		}
	} else {
		fmt.Println(err)
	}

	return token, nil
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// Fetch public keys at a known jwks url and find the public key that corresponds
// to the private key that was used to sign a given jwt token.
// `kid` is a jwt header claim which holds a key identifier, it lets us find the key
// that was used to sign the token in the jwks response.
func getPemCert(token *jwt.Token, jwksURL string) (string, error) {
	var cert = ""
	var jwks = Jwks{}

	op := func() error {
		resp, err := http.Get(jwksURL)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	backOff := backoff.WithMaxRetries(
		backoff.NewExponentialBackOff(),
		3,
	)
	err := backoff.Retry(op, backOff)
	if err != nil {
		return "", microerror.Mask(err)
	}

	x5c := jwks.Keys[0].X5c
	for k, v := range x5c {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + v + "\n-----END CERTIFICATE-----"
			break
		}
	}

	if cert == "" {
		return "", microerror.Maskf(unknownError, "Unable to find a certificate that corresponds to this token.")
	}

	return cert, nil
}
