package commands

import (
	"encoding/base64"
	"fmt"
	"net/http"

	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/microerror"
	gsclientgen "gopkg.in/giantswarm/gsclientgen.v1"

	"github.com/giantswarm/gsctl/config"
)

// loginGiantSwarm executes the authentication logic.
// If the user was logged in before, a logout is performed first.
func loginGiantSwarm(args loginArguments) (loginResult, error) {
	result := loginResult{
		apiEndpoint:        args.apiEndpoint,
		email:              args.email,
		loggedOutBefore:    false,
		endpointSwitched:   false,
		numEndpointsBefore: config.Config.NumEndpoints(),
	}

	endpointBefore := config.Config.SelectedEndpoint
	if result.apiEndpoint != endpointBefore {
		result.endpointSwitched = true
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(args.password))

	// log out if logged in
	if config.Config.Token != "" {
		result.loggedOutBefore = true
		// we deliberately ignore the logout result here
		logout(logoutArguments{
			apiEndpoint: args.apiEndpoint,
			token:       config.Config.Token,
		})
	}

	requestBody := gsclientgen.LoginBodyModel{Password: string(encodedPassword)}
	loginResponse, rawResponse, err := Client.UserLogin(args.email,
		requestBody, loginActivityName)

	if err != nil {
		if rawResponse == nil || rawResponse.Response == nil {
			return result, microerror.Mask(noResponseError)
		}

		if rawResponse.StatusCode == http.StatusForbidden {
			return result, microerror.Mask(accessForbiddenError)
		}

		return result, microerror.Mask(err)
	}

	switch loginResponse.StatusCode {
	case apischema.STATUS_CODE_DATA:
		// successful login
		result.token = loginResponse.Data.Id
		result.email = args.email

		// fetch installation name as alias
		alias, err := getAlias(args.apiEndpoint, "giantswarm", result.token)
		if err != nil {
			return result, microerror.Mask(err)
		}

		result.alias = alias

		if err := config.Config.StoreEndpointAuth(args.apiEndpoint, result.alias, args.email, "giantswarm", result.token); err != nil {
			return result, microerror.Mask(err)
		}
		if err := config.Config.SelectEndpoint(args.apiEndpoint); err != nil {
			return result, microerror.Mask(err)
		}

		// after storing endpoint, get new endpoint count
		result.numEndpointsAfter = config.Config.NumEndpoints()

		return result, nil

	case apischema.STATUS_CODE_RESOURCE_INVALID_CREDENTIALS:
		// bad credentials
		return result, microerror.Mask(invalidCredentialsError)
	case apischema.STATUS_CODE_RESOURCE_NOT_FOUND:
		// user unknown or user/password mismatch
		return result, microerror.Mask(invalidCredentialsError)
	case apischema.STATUS_CODE_WRONG_INPUT:
		// empty password
		return result, microerror.Mask(emptyPasswordError)
	case apischema.STATUS_CODE_ACCOUNT_INACTIVE:
		return result, microerror.Mask(userAccountInactiveError)
	default:
		return result, fmt.Errorf("Unhandled response code: %v", loginResponse.StatusCode)
	}
}
