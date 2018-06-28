package commands

import (
	"math/rand"
	"time"

	"github.com/giantswarm/gsctl/client"
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
	idToken, err := pkce.ParseIdToken(pkceResponse.IDToken)
	if err != nil {
		return loginResult{}, microerror.Maskf(ssoError, "Unable to parse the IDToken")
	}

	// Check if the access token works by fetching the installation's name.
	alias, err := getAlias(args.apiEndpoint, pkceResponse.AccessToken)
	if err != nil {
		return loginResult{}, microerror.Maskf(ssoError, "Unable to fetch installation information. Our api might be experiencing difficulties")
	}

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, idToken.Email, "Bearer", pkceResponse.AccessToken); err != nil {
		return loginResult{}, microerror.Maskf(ssoError, "Error while attempting to store the token in the config file")
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

// getAlias creates a giantswarm API client and tries to fetch the info endpoint.
// If it succeeds it returns the alias for that endpoint.
func getAlias(apiEndpoint string, accessToken string) (string, error) {
	// Create an API client.
	clientConfig := client.Configuration{
		Endpoint:  apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, err := client.NewClient(clientConfig)
	if err != nil {
		return "", err
	}

	// Fetch installation name as alias.
	authHeader := "Bearer " + accessToken
	infoResponse, _, err := apiClient.GetInfo(authHeader, requestIDHeader, loginActivityName, cmdLine)
	if err != nil {
		return "", err
	}

	alias := infoResponse.General.InstallationName

	return alias, nil
}
