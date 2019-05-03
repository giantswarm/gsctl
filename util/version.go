package util

import (
	"github.com/Masterminds/semver"
	"github.com/giantswarm/microerror"
)

// CompareVersions compres to semver version strings v1 and v2.
// Returned result:
// - if v1 is greater than v2: 1
// - if v1 is equal to v2: 0
// if v1 is smaller than v2: -1
func CompareVersions(v1 string, v2 string) (int, error) {
	s1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, microerror.Mask(err)
	}
	s2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return s1.Compare(s2), nil
}

// VersionSortComp is a comparison function usable with the sort.Slice API
func VersionSortComp(v1, v2 string) bool {
	comp, _ := CompareVersions(v1, v2)
	if comp == -1 {
		return true
	}
	return false
}
