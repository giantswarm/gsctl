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
	// allCapabilities is a list of all the capabilities this package knows about.
	allCapabilities []CapabilityDefinition

	// client is an API client the service can use to fetch info.
	clientWrapper *client.Wrapper

	// provider identifies the installation's provide,r e. g. 'aws' or 'azure'.
	provider string
}

// New creates a new configured Service.
func New(provider string, clientWrapper *client.Wrapper) (*Service, error) {
	if provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "Provider must not be empty")
	} else if clientWrapper == nil {
		return nil, microerror.Maskf(invalidConfigError, "Client must not be empty")
	}

	s := &Service{
		provider:      provider,
		clientWrapper: clientWrapper,

		allCapabilities: []CapabilityDefinition{
			Autoscaling,
			AvailabilityZones,
			NodePools,
			HAMasters,
		},
	}

	err := s.initCapabilities()
	if err != nil {
		return nil, microerror.Maskf(couldNotInitializeCapabilities, err.Error())
	}

	return s, nil
}

// initCapabilities adds information taken from the API to our internal capabilities data.
func (s *Service) initCapabilities() error {
	info, err := s.clientWrapper.GetInfo(nil)
	if err != nil {
		return microerror.Maskf(couldNotFetchFeatures, err.Error())
	}

	if info.Payload.Features == nil {
		return nil
	}

	// Enhance feature info for Node Pools, if available.
	if info.Payload.Features.Nodepools != nil {
		NodePools.RequiredReleasePerProvider = []ReleaseProviderPair{
			{
				Provider:       info.Payload.General.Provider,
				ReleaseVersion: semver.MustParse(info.Payload.Features.Nodepools.ReleaseVersionMinimum),
			},
		}
	}

	// Enhance feature info for HA Masters, if available.
	if info.Payload.Features.HaMasters != nil {
		HAMasters.RequiredReleasePerProvider = []ReleaseProviderPair{
			{
				Provider:       info.Payload.General.Provider,
				ReleaseVersion: semver.MustParse(info.Payload.Features.HaMasters.ReleaseVersionMinimum),
			},
		}
	}

	return nil
}

// GetCapabilities returns the list of capabilities that applies to a given release version,
// considering the installation's provider.
func (s *Service) GetCapabilities(releaseVersion string) ([]CapabilityDefinition, error) {
	capabilities := []CapabilityDefinition{}

	// iterate all capabilities and find the ones that apply
	for _, capability := range s.allCapabilities {
		hasCap, err := s.HasCapability(releaseVersion, capability)
		if err != nil {
			return []CapabilityDefinition{}, microerror.Mask(err)
		}
		if hasCap {
			capabilities = append(capabilities, capability)
		}
	}

	return capabilities, nil
}

// HasCapability returns true if the current context (provider, release) provides
// the given capabililty.
func (s *Service) HasCapability(releaseVersion string, capability CapabilityDefinition) (bool, error) {
	ver, err := semver.NewVersion(releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// check which release/provider pair matches ours
	for _, releaseProviderPair := range capability.RequiredReleasePerProvider {
		if s.provider == releaseProviderPair.Provider {
			if !ver.LessThan(releaseProviderPair.ReleaseVersion) {
				return true, nil
			}
		}
	}

	return false, nil
}
