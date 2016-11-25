package commands

// This file defines some variables to be available in all commands

import (
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
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
