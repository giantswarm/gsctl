// Package capabilities provides an service to find out which capabilities/functions
// a tenant cluster on the given installation will provide.
package capabilities

import (
	"github.com/Masterminds/semver"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client"
)

// Service provides methods to get more details on the installation's
// and tenant cluster's capabilities.
type Service struct {
	// Provider identifies the installation's provide,r e. g. 'aws' or 'azure'.
	Provider string

	// Client is an API client the service can use to fetch info.
	ClientWrapper *client.Wrapper
}

// New creates a new configured Service.
func New(provider string, clientWrapper *client.Wrapper) (*Service, error) {
	if provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "Provider must not be empty")
	} else if clientWrapper == nil {
		return nil, microerror.Maskf(invalidConfigError, "Client must not be empty")
	}

	s := &Service{
		Provider:      provider,
		ClientWrapper: clientWrapper,
	}

	err := s.initCapabilities()
	if err != nil {
		return nil, microerror.Maskf(couldNotInitializeCapabilities, err.Error())
	}

	return s, nil
}

// initCapabilities adds information taken from the API to our internal capabilities data.
func (s *Service) initCapabilities() error {
	info, err := s.ClientWrapper.GetInfo(nil)
	if err != nil {
		return microerror.Maskf(couldNotFetchFeatures, err.Error())
	}

	// Enhance feature info for Node Pools, if available.
	if info.Payload.Features.Nodepools != nil {
		NodePools.RequiredReleasePerProvider = []ReleaseProviderPair{
			ReleaseProviderPair{
				Provider:       info.Payload.General.Provider,
				ReleaseVersion: semver.MustParse(info.Payload.Features.Nodepools.ReleaseVersionMinimum),
			},
		}
	}

	return nil
}

// GetCapabilities returns the list of capabilities that applies to a given release version,
// copnsidering the installation's provider.
// This is where some capability definitions get completed using information from the API.
func (s *Service) GetCapabilities(releaseVersion string) ([]*CapabilityDefinition, error) {
	capabilities := []*CapabilityDefinition{}

	// iterate all capabilities and find the ones that apply
	for _, capability := range AllCapabilityDefinitions {
		hasCap, err := s.HasCapability(releaseVersion, capability)
		if err != nil {
			return []*CapabilityDefinition{}, microerror.Mask(err)
		}
		if hasCap {
			capabilities = append(capabilities, capability)
		}
	}

	return capabilities, nil
}

// HasCapability returns true if the current context (provider, release) provides
// the given capabililty.
func (s *Service) HasCapability(releaseVersion string, capability *CapabilityDefinition) (bool, error) {
	ver, err := semver.NewVersion(releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// check which release/provider pair matches ours
	for _, releaseProviderPair := range capability.RequiredReleasePerProvider {
		if s.Provider == releaseProviderPair.Provider {
			if !ver.LessThan(releaseProviderPair.ReleaseVersion) {
				return true, nil
			}
		}
	}

	return false, nil
}
