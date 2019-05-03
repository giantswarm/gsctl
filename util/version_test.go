package util

import (
	"testing"
)

func TestVersionCompare(t *testing.T) {
	var testCases = []struct {
		in  []string
		out int
	}{
		{[]string{"0.1.2", "0.1.3"}, -1},
		{[]string{"0.1.2", "0.1.2"}, 0},
		{[]string{"0.1.3", "0.1.2"}, 1},
		{[]string{"0.1", "0.1.0"}, 0},
		{[]string{"0.1", "0.1.1"}, -1},
		{[]string{"0.2", "0.1.1"}, 1},
	}

	for index, tt := range testCases {
		result, err := CompareVersions(tt.in[0], tt.in[1])
		if err != nil {
			t.Errorf("Test %d: Unexpected error '%s'", index, err)
		}
		if result != tt.out {
			t.Errorf("Test %d: Expected %d, got %d", index, tt.out, result)
		}
	}
}

func TestVersionSortComp(t *testing.T) {
	var testCases = []struct {
		in  []string
		out bool
	}{
		{[]string{"0.1.2", "0.1.3"}, true},
		{[]string{"0.1.2", "0.1.2"}, false},
		{[]string{"0.1.3", "0.1.2"}, false},
	}

	for index, tt := range testCases {
		result := VersionSortComp(tt.in[0], tt.in[1])
		if result != tt.out {
			t.Errorf("Test %d: Expected %v, got %v", index, tt.out, result)
		}
	}
}
