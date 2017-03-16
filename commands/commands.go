package commands

// This file defines some variables to be available in all commands

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen"
)

var (
	// API endpoint flag
	cmdAPIEndpoint string

	// token flag
	cmdToken string

	// verbose flag
	cmdVerbose bool

	// cluster ID flag
	cmdClusterID string

	// description flag
	cmdDescription string

	// TTL (time to live) flag
	cmdTTLDays int

	// Key pair ID flag
	cmdKeypairID string

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	requestIDHeader = randomRequestID()
	cmdLine = getCommandLine()
}

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
	if os.Getenv("GSCTL_DISABLE_CMDLINE_TRACKING") == "" {
		return strings.Join(os.Args, " ")
	}
	return ""
}

// dumpAPIResponse prints details on an API response, useful in case of an error
func dumpAPIResponse(response gsclientgen.APIResponse) {
	output := []string{}
	fmt.Println("API request/response details:")
	output = append(output, fmt.Sprintf("Operation:|%s (%s %s)", response.Operation, response.Method, response.RequestURL))
	output = append(output, fmt.Sprintf("Status:|%s", response.Response.Status))
	output = append(output, fmt.Sprintf("Response body:|%v", response.Payload))
	fmt.Println(columnize.SimpleFormat(output))
}
