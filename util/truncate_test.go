package util

import "testing"

type TruncateTest struct {
	input  string
	length int
	output string
}

var truncateTests = []TruncateTest{
	{"This is a longer string", 10, "This is a…"},
	{"Cut after a blank character", 19, "Cut after a blank …"},
	{"This won't be touched", 30, "This won't be touched"},
}

func TestTruncate(t *testing.T) {
	for i, test := range truncateTests {
		out := Truncate(test.input, test.length, true)
		if out != test.output {
			t.Errorf("#%d: TestTruncate(%s, %d) = '%s'; want '%s'", i, test.input, test.length, out, test.output)
		}
	}
}
