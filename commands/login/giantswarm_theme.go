package login

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client"
)

// loginGiantSwarm executes the authentication logic.
// If the user was logged in before, a logout is performed first.
func loginGiantSwarm(args Arguments) (loginResult, error) {
	result := loginResult{
		apiEndpoint:        args.apiEndpoint,
		email:              args.email,
		provider:           "",
		loggedOutBefore:    false,
		endpointSwitched:   false,
		numEndpointsBefore: config.Config.NumEndpoints(),
	}

	endpointBefore := config.Config.SelectedEndpoint
	if result.apiEndpoint != endpointBefore {
		result.endpointSwitched = true
	}

	clientV2, err := client.NewWithConfig(args.apiEndpoint, "")
	if err != nil {
		return result, microerror.Mask(err)
	}

	ap := clientV2.DefaultAuxiliaryParams()
	ap.ActivityName = loginActivityName

	// log out if logged in
	if config.Config.Token != "" {
		if args.verbose {
			fmt.Println(color.WhiteString("Logging out using a a previously stored token"))
		}

		result.loggedOutBefore = true
		// we deliberately ignore the logout result here
		clientV2.DeleteAuthToken(config.Config.Token, ap)
	}

	if args.verbose {
		fmt.Println(color.WhiteString("Submitting API call to create an authentication token with email '%s'", args.email))
	}

	response, err := clientV2.CreateAuthToken(args.email, args.password, ap)
	if err != nil {
		return result, err
	}

	// handle success

	result.token = response.Payload.AuthToken
	result.email = args.email

	// fetch installation name as alias
	if args.verbose {
		fmt.Println(color.WhiteString("Fetching installation details"))
	}

	installationInfo, err := getInstallationInfo(args.apiEndpoint, "giantswarm", result.token)
	if err != nil {
		return result, microerror.Mask(err)
	}
	result.alias = installationInfo.InstallationName
	result.provider = installationInfo.Provider

	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, result.alias, result.provider, args.email, "giantswarm", result.token, ""); err != nil {
		return result, microerror.Mask(err)
	}
	if err := config.Config.SelectEndpoint(args.apiEndpoint); err != nil {
		return result, microerror.Mask(err)
	}

	// after storing endpoint, get new endpoint count
	result.numEndpointsAfter = config.Config.NumEndpoints()

	return result, nil
}
