package commands

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/oidc"
	"github.com/giantswarm/microerror"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func loginSSO(args loginArguments) (loginResult, error) {
	numEndpointsBefore := config.Config.NumEndpoints()

	pkceResponse, err := oidc.RunPKCE()
	if err != nil {
		if args.verbose {
			fmt.Println(color.WhiteString("Attempt to run the OAuth2 PKCE workflow with a local callback HTTP server failed."))
		}
		return loginResult{}, microerror.Maskf(ssoError, pkceResponse.ErrorDescription)
	}

	// Try to parse the ID Token.
	idToken, err := oidc.ParseIDToken(pkceResponse.IDToken)
	if err != nil {
		return loginResult{}, microerror.Mask(err)
	}

	// Check if the access token works by fetching the installation's name.
	alias, err := getAlias(args.apiEndpoint, "Bearer", pkceResponse.AccessToken)
	if err != nil {
		if args.verbose {
			fmt.Println(color.WhiteString("Attempt to use new token against the API failed."))
			if cErr, ok := err.(*clienterror.APIError); ok {
				fmt.Println(color.WhiteString("Underlying error details: %s", cErr.OriginalError.Error()))
			} else {
				fmt.Println(color.WhiteString("Error details: %s", err.Error()))
			}
		}

		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusForbidden {
				return loginResult{}, microerror.Mask(accessForbiddenError)
			}

			return loginResult{}, clientErr
		}

		return loginResult{}, microerror.Maskf(ssoError, err.Error())
	}

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, idToken.Email, "Bearer", pkceResponse.AccessToken, pkceResponse.RefreshToken); err != nil {
		if args.verbose {
			fmt.Println(color.WhiteString("Attempt to store our authentication data with the endpoint in the configuration failed."))
			fmt.Println(color.WhiteString("Error details: %s", err.Error()))
		}
		return loginResult{}, microerror.Maskf(ssoError, "Error while attempting to store the token in the config file")
	}
	if err := config.Config.SelectEndpoint(args.apiEndpoint); err != nil {
		return loginResult{}, microerror.Mask(err)
	}

	result := loginResult{
		apiEndpoint:        args.apiEndpoint,
		email:              idToken.Email,
		endpointSwitched:   (config.Config.SelectedEndpoint != args.apiEndpoint),
		loggedOutBefore:    false,
		alias:              alias,
		token:              pkceResponse.AccessToken,
		numEndpointsBefore: numEndpointsBefore,
		numEndpointsAfter:  config.Config.NumEndpoints(),
	}

	return result, nil
}
