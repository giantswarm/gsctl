package commands

// This file defines some variables to be available in all commands

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen"

	"github.com/fatih/color"
)

var (
	// API endpoint flag
	cmdAPIEndpoint string

	// token flag
	cmdToken string

	// configuration path to use temporarily
	cmdConfigDirPath string

	// verbose flag
	cmdVerbose bool

	// cluster ID flag
	cmdClusterID string

	// description flag
	cmdDescription string

	// TTL (time to live) flag
	cmdTTLDays int

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string

	// errors for common purposes
	errNotLoggedIn           = "user not logged in"
	errConflictingFlags      = "conflicting flags used"
	errClusterIDNotSpecified = "cluster ID not specified"
	errClusterNotFound       = "the cluster specified could not be found"
	errInternalServerError   = "an internal server error occurred"
)

// APIError is an error type we use for errors generated after API requests
type APIError struct {
	// msg is the error message
	msg string
	// APIResponse is the response we got from the API
	APIResponse gsclientgen.APIResponse
}

func (e APIError) Error() string {
	return e.msg
}

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
	if response.Response == nil {
		fmt.Println("No response received")
	} else {
		output := []string{}
		fmt.Println("API request/response details:")
		output = append(output, fmt.Sprintf("Operation:|%s (%s %s)", response.Operation, response.Method, response.RequestURL))
		output = append(output, fmt.Sprintf("Status:|%s", response.Response.Status))
		output = append(output, fmt.Sprintf("Response body:|%v", string(response.Payload)))
		fmt.Println(columnize.SimpleFormat(output))
	}
}

// askForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", color.YellowString(s))

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
