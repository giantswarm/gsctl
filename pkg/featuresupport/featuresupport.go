package featuresupport

import (
	"strings"

	"github.com/Masterminds/semver"
)

type Feature struct {
	Providers []Provider
}

// IsSupported checks to see if a feature is supported by a specific provider,
// for a specific release version.
func (f *Feature) IsSupported(provider string, version string) bool {
	p := f.getProviderWithName(provider)
	if p == nil {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}
	requiredVersion, err := semver.NewVersion(p.RequiredVersion)
	if err != nil {
		return false
	}

	return v.Compare(requiredVersion) >= 0
}

// RequiredVersion returns a feature's required version for a specific provider.
func (f *Feature) RequiredVersion(provider string) string {
	p := f.getProviderWithName(provider)
	if p == nil {
		return "0.0.1"
	}

	return p.RequiredVersion
}

func (f *Feature) getProviderWithName(p string) *Provider {
	p = strings.ToLower(p)

	for _, provider := range f.Providers {
		if strings.ToLower(provider.Name) == p {
			return &provider
		}
	}

	return nil
}
