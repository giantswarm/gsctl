package oidc

import (
	"encoding/json"
	"net/http"

	"github.com/cenkalti/backoff"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/giantswarm/microerror"
)

const (
	jwksURL = "https://giantswarm.eu.auth0.com/.well-known/jwks.json"
)

// IDToken is our custom representation of the details of a JWT we care about
type IDToken struct {
	// Email claim
	Email string
}

// ParseIDToken takes a jwt token and returns an IDToken, which is just a custom
// struct with only the email claim in it. Since that is all that gsctl cares about
// for now.
func ParseIDToken(tokenString string) (token *IDToken, err error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	t, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		cert, err := getPemCert(token, jwksURL)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))

		return result, nil
	})
	if err != nil {
		// handle some validation errors specifically
		valErr, valErrOK := err.(*jwt.ValidationError)
		if valErrOK && valErr.Errors == jwt.ValidationErrorIssuedAt {
			return nil, microerror.Maskf(tokenIssuedAtError, valErr.Error())
		}

		return nil, microerror.Maskf(tokenInvalidError, err.Error())
	}

	if !t.Valid {
		return nil, microerror.Mask(tokenInvalidError)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, microerror.Mask(tokenInvalidError)
	}

	if claims == nil {
		return nil, microerror.Mask(tokenInvalidError)
	}

	resultToken := &IDToken{}

	if email, ok := claims["email"]; ok {
		resultToken.Email = email.(string)
	}

	return resultToken, nil
}

// Jwks holds JSON web keys.
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

// JSONWebKeys represents one JWS web key.
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
		return "", microerror.Maskf(authorizationError, "Unable to find a certificate that corresponds to this token.")
	}

	return cert, nil
}