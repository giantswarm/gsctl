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

type cpuDefinition struct {
	Cores int `yaml:"cores,omitempty"`
}

type memoryDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

type storageDefinition struct {
	SizeGB float32 `yaml:"size_gb,omitempty"`
}

type awsSpecificDefinition struct {
	InstanceType string `yaml:"instance_type,omitempty"`
}

type azureSpecificDefinition struct {
	VMSize string `yaml:"vm_size,omitempty"`
}

type nodeDefinition struct {
	Memory  memoryDefinition        `yaml:"memory,omitempty"`
	CPU     cpuDefinition           `yaml:"cpu,omitempty"`
	Storage storageDefinition       `yaml:"storage,omitempty"`
	Labels  map[string]string       `yaml:"labels,omitempty"`
	AWS     awsSpecificDefinition   `yaml:"aws,omitempty"`
	Azure   azureSpecificDefinition `yaml:"azure,omitempty"`
}

type clusterDefinition struct {
	Name              string            `yaml:"name,omitempty"`
	Owner             string            `yaml:"owner,omitempty"`
	ReleaseVersion    string            `yaml:"release_version,omitempty"`
	AvailabilityZones int               `yaml:"availability_zones,omitempty"`
	Scaling           scalingDefinition `yaml:"scaling,omitempty"`
	Workers           []nodeDefinition  `yaml:"workers,omitempty"`
}

type scalingDefinition struct {
	Min int64 `yaml:"min,omitempty"`
	Max int64 `yaml:"max,omitempty"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
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
