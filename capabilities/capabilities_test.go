package capabilities

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
)

var capabilityTests = []struct {
	in  ReleaseProviderPair
	out []CapabilityDefinition
}{
	{
		ReleaseProviderPair{
			Provider:       "aws",
			ReleaseVersion: semver.MustParse("1.2.3"),
		},
		[]CapabilityDefinition{},
	},
	{
		ReleaseProviderPair{
			Provider:       "aws",
			ReleaseVersion: semver.MustParse("6.4.0"),
		},
		[]CapabilityDefinition{Autoscaling},
	},
	{
		ReleaseProviderPair{
			Provider:       "aws",
			ReleaseVersion: semver.MustParse("9.0.0"),
		},
		[]CapabilityDefinition{Autoscaling, NodePools},
	},
	{
		ReleaseProviderPair{
			Provider:       "aws",
			ReleaseVersion: semver.MustParse("9.1.2"),
		},
		[]CapabilityDefinition{Autoscaling, NodePools},
	},
	{
		ReleaseProviderPair{
			Provider:       "kvm",
			ReleaseVersion: semver.MustParse("9.1.2"),
		},
		[]CapabilityDefinition{},
	},
}

func TestGetCapabilities(t *testing.T) {
	for index, tt := range capabilityTests {
		cap, err := GetCapabilities(tt.in.Provider, tt.in.ReleaseVersion)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(cap, tt.out) {
			t.Errorf("Test %d: Expected %#v but got %#v", index, tt.out, cap)
		}
	}
}
