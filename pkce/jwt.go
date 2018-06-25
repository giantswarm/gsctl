package pkce

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/giantswarm/microerror"
)

const (
	jwksURL = "https://giantswarm.eu.auth0.com/.well-known/jwks.json"
)

type IDToken struct {
	Email string
}

// ParseIdToken takes a jwt token and returns an IDToken, which is just a custom
// struct with only the email claim in it. Since that is all that gsctl cares about
// for now.
func ParseIdToken(tokenString string) (token IDToken, err error) {
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
		return "", microerror.Maskf(authorizationError, "Unable to find a certificate that corresponds to this token.")
	}

	return cert, nil
}
