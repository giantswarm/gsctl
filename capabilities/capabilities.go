package capabilities

import "github.com/Masterminds/semver"

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

	// AvailabilityZones is the capability to spread the worker nodes of a tenant
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
	// Details get completed with API data, if the feature is available.
	NodePools = CapabilityDefinition{
		Name: "NodePools",
	}

	// HAMasters provides details about the high availability masters feature.
	HAMasters = CapabilityDefinition{
		Name: "HAMasters",
	}
)
