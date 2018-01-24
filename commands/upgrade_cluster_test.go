package commands

import (
	"testing"
)

var successorVersionTests = []struct {
	myVersion        string
	allVersions      []string
	successorVersion string
}{
	{
		"1.2.3",
		[]string{"1.2.3", "1.2.4", "1.2.5", "10.2.3", "2.0.0", "0.1.0"},
		"1.2.4",
	},
	// none of the versions in the slice is higher
	{
		"4.5.2",
		[]string{"3.2.1", "0.5.1", "0.5.0", "0.6.0", "4.5.2"},
		"",
	},
}

func TestSuccessorReleaseVersion(t *testing.T) {
	for i, tc := range successorVersionTests {
		v := successorReleaseVersion(tc.myVersion, tc.allVersions)
		if v != tc.successorVersion {
			t.Errorf("%d. Expected %s, got %s", i, tc.successorVersion, v)
		}
	}
}
