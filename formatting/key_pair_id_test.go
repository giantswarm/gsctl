package formatting

import (
	"strconv"
	"testing"
)

type KeypairIDTest struct {
	input  string
	output string
}

func TestCleanKeypairID(t *testing.T) {
	testCases := []KeypairIDTest{
		{"a1:b2:c3:d4:e5:f6:g7:00", "a1b2c3d4e5f6g700"},
	}

	for i, test := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := CleanKeypairID(test.input)
			if out != test.output {
				t.Errorf("#%d: CleanKeypairID(%s) = '%s'; want '%s'", i, test.input, out, test.output)
			}
		})
	}
}
