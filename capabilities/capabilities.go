// Package capabilities provides an API for capability detection.
package capabilities

import (
	"github.com/Masterminds/semver"
	"github.com/giantswarm/microerror"
)

var (
	// Autoscaling is the capability to scale tenant clusters automatically.
	Autoscaling = CapabilityDefinition{
		Name: "Autoscaling",
		RequiredReleasePerProvider: []ReleaseProviderPair{
			ReleaseProviderPair{
				Provider:       "aws",
				ReleaseVersion: semver.MustParse("6.3"),
			},
		},
	}

	// AvailabilityZones si the capability to spread the worker nodes of a tenant
	// cluster over multiple availability zones.
	AvailabilityZones = CapabilityDefinition{
		Name: "AvailabilityZones",
		RequiredReleasePerProvider: []ReleaseProviderPair{
			ReleaseProviderPair{
				Provider:       "aws",
				ReleaseVersion: semver.MustParse("6.1"),
			},
		},
	}

	// NodePools is the capabilitiy to group tenant cluster workers logically.
	NodePools = CapabilityDefinition{
		Name: "NodePools",
		RequiredReleasePerProvider: []ReleaseProviderPair{
			ReleaseProviderPair{
				Provider:       "aws",
				ReleaseVersion: semver.MustParse("9"), // TODO: fix once the node pools release version is defined.
			},
		},
	}

	// AllCapabilityDefinitions contains all the capabilities
	AllCapabilityDefinitions = []CapabilityDefinition{
		Autoscaling,
		AvailabilityZones,
		NodePools,
	}
)

// CapabilityDefinition is the type we use to describe the conditions that have to be met
// so that we can assume a certain capability on the installation or API side.
type CapabilityDefinition struct {
	Name        string
	Description string
	// RequiredReleasePerProvider holds the combination(s) of provider and
	// release version which have to be fulfilled so we assume a capability.
	RequiredReleasePerProvider []ReleaseProviderPair
}

// ReleaseProviderPair is a combination of a providr ('aws', 'azure', 'kvm) and
// a release version number.
type ReleaseProviderPair struct {
	Provider       string
	ReleaseVersion *semver.Version
}

// GetCapabilities returns the capabilities available in the current context
func GetCapabilities(provider, releaseVersion string) ([]CapabilityDefinition, error) {
	cap := []CapabilityDefinition{}

	// iterate all capabilities and find the ones that apply
	for _, capability := range AllCapabilityDefinitions {
		hasCap, err := HasCapability(provider, releaseVersion, capability)
		if err != nil {
			return []CapabilityDefinition{}, microerror.Mask(err)
		}
		if hasCap {
			cap = append(cap, capability)
		}
	}

	return cap, nil
}

// HasCapability returns true if the current context (provider, release) provides
// the given capabililty.
func HasCapability(provider, releaseVersion string, capability CapabilityDefinition) (bool, error) {
	ver, err := semver.NewVersion(releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// check which release/provider pair matches ours
	for _, releaseProviderPair := range capability.RequiredReleasePerProvider {
		if provider == releaseProviderPair.Provider {
			if !ver.LessThan(releaseProviderPair.ReleaseVersion) {
				return true, nil
			}
		}
	}

	return false, nil
}
