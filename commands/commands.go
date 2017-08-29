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

	// certificate_organizations flag
	cmdCertificateOrganizations string

	// cluster ID flag
	cmdClusterID string

	// cn_prefox flag
	cmdCNPrefix string

	// description flag
	cmdDescription string

	// TTL (time to live) flag
	cmdTTLDays int

	// force flag. if set, no prompt should be displayed.
	cmdForce bool

	// number of CPUs per worker as required via flag on execution
	cmdWorkerNumCPUs int

	// RAM size per worker node in GB per worker as required via flag on execution
	cmdWorkerMemorySizeGB float32

	// Local storage per worker node in GB per worker as required via flag on execution
	cmdWorkerStorageSizeGB float32

	randomStringCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	requestIDHeader string
	cmdLine         string
)

// APIError is an error type we use for errors generated after API requests
type APIError struct {
	// msg is the error message
	msg string
	// APIResponse is the response we got from the API
	APIResponse gsclientgen.APIResponse
}

type cpuDefinition struct {
	Cores int `yaml:"cores,omitempty"`
}

type memoryDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

type storageDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

type nodeDefinition struct {
	Memory  memoryDefinition  `yaml:"memory,omitempty"`
	CPU     cpuDefinition     `yaml:"cpu,omitempty"`
	Storage storageDefinition `yaml:"storage,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

type clusterDefinition struct {
	Name              string           `yaml:"name,omitempty"`
	Owner             string           `yaml:"owner,omitempty"`
	KubernetesVersion string           `yaml:"kubernetes_version,omitempty"`
	Workers           []nodeDefinition `yaml:"workers,omitempty"`
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
