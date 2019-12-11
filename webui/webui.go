// Package webui provides a method to find the Web UI (happa) URL, based on the installations's API endpoint URL.
package webui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	apiStandardHostName = "api"

	webUIStandardHostName = "happa"
)

// BaseURL returns the web UI base URL derived from an API URL.
func BaseURL(apiEndpoint string) (string, error) {
	parsed, err := url.Parse(apiEndpoint)
	if err != nil {
		return "", microerror.Mask(err)
	}

	hostNameParts := strings.Split(parsed.Hostname(), ".")

	// replace first part with 'happa' if it is 'api'
	if len(hostNameParts) < 2 {
		return "", microerror.Mask(unsupportedHostNameError)
	}

	if hostNameParts[0] != apiStandardHostName {
		return "", microerror.Maskf(unsupportedHostNameError, "host name must start with 'api'")
	}

	hostNameParts[0] = webUIStandardHostName

	webUIFullHostName := strings.Join(hostNameParts, ".")

	return "https://" + webUIFullHostName, nil
}

// ClusterDetailsURL returns the URL of a cluster details page in the web UI.
func ClusterDetailsURL(apiEndpoint string, clusterID string, organization string) (string, error) {
	if clusterID == "" {
		return "", microerror.Maskf(missingArgumentError, "cluster ID must be given")
	}
	if organization == "" {
		return "", microerror.Maskf(missingArgumentError, "organization ID must be given")
	}

	baseURL, err := BaseURL(apiEndpoint)
	if err != nil {
		return "", microerror.Mask(err)
	}

	url := fmt.Sprintf("%s/organizations/%s/clusters/%s", baseURL, organization, clusterID)

	return url, nil
}
