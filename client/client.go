package client

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	apiclient "github.com/giantswarm/gsclientgen/client"
	"github.com/giantswarm/microerror"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	rootcerts "github.com/hashicorp/go-rootcerts"
)

// Configuration is the configuration to be used when creating a new API client
type Configuration struct {
	// Endpoint is the base URL of the API
	Endpoint string
	// Usergent is the string we pass with the HTTP User-Agent header
	UserAgent string
}

// GenericResponse allows to access details of a generic API response (mostly error messages).
type GenericResponse struct {
	Code    string
	Message string
}

type roundTripperWithUserAgent struct {
	inner http.RoundTripper
	Agent string
}

// RoundTrip overwrites the http.RoundTripper.RoundTrip function to add our
// User-Agent HTTP header.
func (rtwua *roundTripperWithUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", rtwua.Agent)
	return rtwua.inner.RoundTrip(r)
}

// setUserAgent sets the User-Agent header value for subsequent requests
// made using this transport
func setUserAgent(inner http.RoundTripper, userAgent string) http.RoundTripper {
	return &roundTripperWithUserAgent{
		inner: inner,
		Agent: userAgent,
	}
}

// NewClient allows to create a new API client
// with specific configuration
func NewClient(clientConfig Configuration) (*apiclient.Gsclientgen, error) {

	if clientConfig.Endpoint == "" {
		return nil, microerror.Mask(endpointNotSpecifiedError)
	}

	// parse endpoint URL
	u, err := url.Parse(clientConfig.Endpoint)
	if err != nil {
		return nil, microerror.Mask(endpointInvalidError)
	}

	// set up client TLS so that custom CAs are accepted.
	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		return nil, microerror.Mask(rootCertsErr)
	}

	// Pass TLS and http proxy
	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	transport.Transport = setUserAgent(transport.Transport, clientConfig.UserAgent)

	return apiclient.New(transport, strfmt.Default), nil
}

// ParseGenericResponse parses the standard code, message response document into
// a struct with the fields Code and Message.
func ParseGenericResponse(jsonBody []byte) (GenericResponse, error) {
	var response = GenericResponse{}
	err := json.Unmarshal(jsonBody, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
