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

	"github.com/fatih/color"
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

// askForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
// Thanks to https://gist.github.com/m4ng0squ4sh/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s %s: ", s, color.YellowString("[y/n]"))

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
