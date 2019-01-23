package util

import (
	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

func CompareVersions(v1 string, v2 string) (int, error) {
	s1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, microerror.Mask(err)
	}
	s2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return s1.Compare(*s2), nil
}
