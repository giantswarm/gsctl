package capabilities

import "github.com/Masterminds/semver"

// CapabilityDefinition is the type we use to describe the conditions that have to be met
// so that we can assume a certain capability on the installation or API side.
type CapabilityDefinition struct {
	// Name is a user friendly name we use here in gsctl.
	Name string

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
