package formatting

import (
	"strconv"
	"testing"
)

func TestCleanKeypairID(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"a1:b2:c3:d4:e5:f6:g7:00", "a1b2c3d4e5f6g700"},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := CleanKeypairID(tc.input)
			if out != tc.output {
				t.Errorf("#%d: CleanKeypairID(%s) = '%s'; want '%s'", i, tc.input, out, tc.output)
			}
		})
	}
}
