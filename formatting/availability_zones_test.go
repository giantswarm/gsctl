package formatting

import "testing"

var flagtests = []struct {
	in  []string
	out string
}{
	{[]string{"foo1-c"}, "C"},
	{[]string{"foo1-b", "foo1-a"}, "A,B"},
}

func TestAvailabilityZonesList(t *testing.T) {
	for _, tt := range flagtests {
		s := AvailabilityZonesList(tt.in)
		if s != tt.out {
			t.Errorf("got %q, want %q", s, tt.out)
		}
	}
}
