package formatting

import "testing"

type KeypairIDTest struct {
	input  string
	output string
}

var keypairIDTests = []KeypairIDTest{
	{"a1:b2:c3:d4:e5:f6:g7:00", "a1b2c3d4e5f6g700"},
}

func TestCleanKeypairID(t *testing.T) {
	for i, test := range keypairIDTests {
		out := CleanKeypairID(test.input)
		if out != test.output {
			t.Errorf("#%d: CleanKeypairID(%s) = '%s'; want '%s'", i, test.input, out, test.output)
		}
	}
}
