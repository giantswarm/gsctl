package commands

import (
	"math/rand"
	"time"

	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/pkce"
	"github.com/giantswarm/microerror"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func loginSSO(args loginArguments) (loginResult, error) {
	numEndpointsBefore := config.Config.NumEndpoints()

	pkceResponse, err := pkce.Run()
	if err != nil {
		return loginResult{}, microerror.Maskf(ssoError, pkceResponse.ErrorDescription)
	}

	// Try to parse the ID Token.
	idToken, err := pkce.ParseIDToken(pkceResponse.IDToken)
	if err != nil {
		return loginResult{}, microerror.Mask(err)
	}

	// Check if the access token works by fetching the installation's name.
	alias, err := getAlias(args.apiEndpoint, "Bearer", pkceResponse.AccessToken)
	if err != nil {
		return loginResult{}, microerror.Maskf(ssoError, "Unable to fetch installation information. Our api might be experiencing difficulties")
	}

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, idToken.Email, "Bearer", pkceResponse.AccessToken); err != nil {
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
