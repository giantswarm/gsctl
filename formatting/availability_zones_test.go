package formatting

import (
	"strconv"
	"testing"
)

func TestAvailabilityZonesList(t *testing.T) {
	var testCases = []struct {
		in  []string
		out string
	}{
		{[]string{"foo1-c"}, "C"},
		{[]string{"foo1-b", "foo1-a"}, "A,B"},
	}

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s := AvailabilityZonesList(tt.in)
			if s != tt.out {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}
