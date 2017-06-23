package commands

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
	microerror "github.com/giantswarm/microkit/error"
	rootcerts "github.com/hashicorp/go-rootcerts"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

var (
	// PingCommand is the "ping" CLI command
	PingCommand = &cobra.Command{
		Use:   "ping",
		Short: "Check API connection",
		Long:  `Tests the connection to the API`,
		Run:   runPingCommand,
	}
)

func init() {
	RootCommand.AddCommand(PingCommand)
}

// runPingCommand executes the ping() function
// and prints output in a user-friendly way
func runPingCommand(cmd *cobra.Command, args []string) {
	duration, err := ping(cmdAPIEndpoint)
	if err != nil {
		fmt.Println(color.RedString("Could not reach API"))
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(color.GreenString("API connection is fine"))
	fmt.Printf("Ping took %v\n", duration)
}

// ping checks the API connection and returns
// duration (in case of success) and error (in case of failure)
func ping(endpointURL string) (time.Duration, error) {
	var duration time.Duration

	// create URI
	u, err := url.Parse(endpointURL)
	if err != nil {
		return duration, microerror.MaskAny(err)
	}
	u, err = u.Parse("/v1/ping")
	if err != nil {
		return duration, microerror.MaskAny(err)
	}

	// create client and request
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return duration, microerror.MaskAny(err)
	}
	request.Header.Set("User-Agent", config.UserAgent())

	// create client
	tlsConfig := &tls.Config{}
	rootCertsErr := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv("GSCTL_CAFILE"),
		CAPath: os.Getenv("GSCTL_CAPATH"),
	})
	if rootCertsErr != nil {
		return duration, microerror.MaskAny(rootCertsErr)
	}
	t := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}
	pingClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}

	start := time.Now()
	resp, err := pingClient.Do(request)
	if err != nil {
		return duration, microerror.MaskAny(err)
	}
	defer resp.Body.Close()

	duration = time.Since(start)
	if resp.StatusCode != http.StatusOK {
		// TODO: return typed error
		return duration, fmt.Errorf("bad status code %d", resp.StatusCode)
	}

	return duration, nil
}
