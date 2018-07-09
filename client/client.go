package client

import (
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	// DefaultTimeout is the standard request timeout applied if not specified
	DefaultTimeout = 60 * time.Second

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	requestIDHeader = randomRequestID()
	cmdLine = getCommandLine()
}

// Configuration is the configuration to be used for the old client
// based on gsclientgen.v1
type Configuration struct {
	// AuthHeader is the header we should use to make API calls
	AuthHeader string
	// Endpoint is the base URL of the API
	Endpoint string
	// Timeout is the maximum time to wait for API requests to succeed
	Timeout time.Duration
	// UserAgent
	UserAgent string
}

// NewClientV2
// func NewClientV2(clientConfig Configuration) (*apiclient.Gsclientgen, error) {
// 	return apiclient.New()
// }

// randomRequestID returns a new request ID
func randomRequestID() string {
	size := 14
	b := make([]rune, size)
	for i := range b {
		b[i] = randomStringCharset[rand.Intn(len(randomStringCharset))]
	}
	return string(b)
}

// getCommandLine returns the command line that has been called
func getCommandLine() string {
	if os.Getenv("GSCTL_DISABLE_CMDLINE_TRACKING") != "" {
		return ""
	}
	args := redactPasswordArgs(os.Args)
	return strings.Join(args, " ")
}

// redactPasswordArgs replaces password in an arguments slice
// with "REDACTED"
func redactPasswordArgs(args []string) []string {
	for index, arg := range args {
		if strings.HasPrefix(arg, "--password=") {
			args[index] = "--password=REDACTED"
		} else if arg == "--password" {
			args[index+1] = "REDACTED"
		} else if len(args) > 1 && args[1] == "login" {
			// this will explicitly only apply to the login command
			if strings.HasPrefix(arg, "-p=") {
				args[index] = "-p=REDACTED"
			} else if arg == "-p" {
				args[index+1] = "REDACTED"
			}
		}
	}
	return args
}

// DeleteAuthToken calls the deleteAuthToken operation in the latest client
//func (c *APIClient) DeleteAuthToken() (response, error) {
//params := apiOperations.NewGetUserOrganizationsParams()
//params.SetTimeout(1 * time.Second)

//_, err := apiClient.Operations.GetUserOrganizations(params)
//}
