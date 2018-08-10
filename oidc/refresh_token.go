package oidc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/giantswarm/microerror"
)

// RefreshResponse represents the result we get when we use a refersh
// token to get a new access token.
type RefreshResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        string `json:"expires_in"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RefreshRequest represents the request that the token refresh endpoint expects
// in the JSON body. It gets marshalled to JSON.
type RefreshRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken performs a POST call to the auth0 token endpoint with a refresh token
// and returns a RefreshToken response, which includes a new access token.
func RefreshToken(refreshToken string) (refreshResponse RefreshResponse, err error) {
	payload := strings.NewReader(fmt.Sprintf(`{
    "grant_type":"refresh_token",
    "client_id": "%s",
    "refresh_token": "%s"
  }`, clientID, refreshToken))

	req, err := http.NewRequest("POST", tokenURL, payload)
	if err != nil {
		return refreshResponse, microerror.Maskf(refreshError, "Unable to construct POST request for Auth0")
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return refreshResponse, microerror.Maskf(refreshError, "Unable to perform POST request to Auth0")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return refreshResponse, microerror.Maskf(refreshError, "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later")
	}

	json.Unmarshal(body, &refreshResponse)

	// This is a real error from Auth0, in this case we have Error and ErrorDescription
	// set by what Auth0 sent us.
	if refreshResponse.Error != "" {
		return refreshResponse, microerror.Maskf(refreshError, refreshResponse.ErrorDescription)
	}

	if res.StatusCode != 200 {
		return refreshResponse, microerror.Maskf(refreshError, "Got an unparseable error from Auth0. Possibly the Auth0 service is down. Try again later")
	}

	return refreshResponse, nil
}
