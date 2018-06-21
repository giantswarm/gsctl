package commands

import (
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
	"strings"
	"time"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var (
	// SSOCommand performs the "sso" function
	SSOCommand = &cobra.Command{
		Use:    "sso",
		Short:  "Single Sign on for Admins",
		Long:   `Prints a list of all clusters you have access to`,
		PreRun: ssoPreRunOutput,
		Run:    ssoRunOutput,
	}
)

const (
	activityName = "sso"
	clientID     = "zQiFLUnrTFQwrybYzeY53hWWfhOKWRAU"
	redirectURI  = "http://localhost:8085"
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

func ssoPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
}

func ssoRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultSSOArguments()

	// Start a local webserver that auth0 can redirect to with the code
	// that we will then exchange for the actual id token.
	// Credit: https://medium.com/@int128/shutdown-http-server-by-endpoint-in-go-2a0e2d7f9b8c
	codeCh := make(chan string)
	callbackServer := startCallbackServer(codeCh)

	// Construct and open the authorization url.
	// 		1. Generate and store a random codeVerifier.
	codeVerifier := base64URLEncode(fmt.Sprint(rand.Int31()))

	// 		2. Using the codeVerifier, generate a sha256 hashed codeChallenge that
	//    	 will be sent in the authorization request.
	codeChallenge := sha256.Sum256([]byte(codeVerifier))

	// 		3. Use the codeChallenge to make a authorizationURL
	authorizationURL := authorizationURL(string(codeChallenge[:]))

	// Open the authorization url in the users browser, which will eventually
	// redirect the user to the local webserver created above.
	open.Run(authorizationURL)

	// Block until we recieve a code from auth0. (In other words, localhost:8085
	// is hit thanks to auth0's redirect by the users browser with /?code=XXXXXXXX)
	var code string
	select {
	case code = <-codeCh:
		// Code recieved, shutdown.
		callbackServer.Shutdown(context.Background())
	}

	// We now have the 'code' which we can then finally exchange
	// for a real id token by doing a final request to Auth0 and passing the code
	// along with the codeVerifier we made at the start.
	tokenResponse, err := getToken(code, codeVerifier)
	if err != nil {
		panic(err)
	}

	// Check if the token works by fetching the installation's name
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		panic(err)
	}

	// fetch installation name as alias
	authHeader := "Bearer " + tokenResponse.IDToken
	infoResponse, _, infoErr := apiClient.GetInfo(authHeader, requestIDHeader, loginActivityName, cmdLine)
	if infoErr != nil {
		panic(err)
	}

	alias := infoResponse.General.InstallationName
	fmt.Println(infoResponse)
	fmt.Println(alias)

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, "TOKEN", "Bearer", tokenResponse.AccessToken); err != nil {
		panic(err)
	}

	fmt.Println(tokenResponse.AccessToken)
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
// Once it recieves this code it will pass it along to the given channel.
func startCallbackServer(codeCh chan string) http.Server {
	m := http.NewServeMux()
	s := http.Server{Addr: ":8085", Handler: m}
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Authorization complete, you may now close this window."))
		// Send query parameter to the channel
		codeCh <- r.URL.Query().Get("code")
	})

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return s
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
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
		panic(err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &tokenResponse)

	return tokenResponse, nil
}
