package login

import (
	"github.com/coreos/go-oidc"
	oidc2 "github.com/giantswarm/gsctl/commands/login/oidc"
	"github.com/giantswarm/microerror"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/afero"
)

const (
	authCallbackURL  = "http://localhost:8085"
	authCallbackPath = "/oauth/callback"
)

var (
	authScopes = []string{oidc.ScopeOpenID, "profile", "email", "groups", "offline_access"}
)

func loginOIDC(args Arguments) (loginResult, error) {
	result := loginResult{}

	i, err := oidc2.NewInstallation("ginger.eu-west-1.aws.gigantic.io", afero.NewOsFs())
	if err != nil {
		return result, microerror.Mask(err)
	}

	authURL := authCallbackURL + authCallbackPath
	err = i.NewAuthenticator(authURL, authScopes)
	if err != nil {
		return result, microerror.Mask(err)
	}

	i.Authenticator.State = oidc2.GenerateState()
	aURL := i.Authenticator.GetAuthURL()
	// Open the authorization url in the user's browser, which will eventually
	// redirect the user to the local webserver we'll create next.
	err = open.Run(aURL)
	if err != nil {
		return result, microerror.Mask(err)
	}

	p, err := oidc2.StartCallbackServer("8085", authCallbackPath, i.Authenticator.HandleCallback)
	if err != nil {
		return result, microerror.Mask(err)
	}

	authResult, ok := p.(oidc2.UserInfo)
	if !ok {
		// TODO: Throw an actual error
		return result, microerror.New("Cannot deserialize authentication result")
	}

	// Store kubeconfig
	err = i.StoreCredentials(&authResult)
	if err != nil {
		return result, microerror.Mask(err)
	}

	result = loginResult{
		apiEndpoint:     i.BaseURL,
		email:           authResult.Email,
		loggedOutBefore: false,
		alias:           i.Alias,
		provider:        i.Provider,
		token:           authResult.IDToken,
	}

	return result, nil
}
