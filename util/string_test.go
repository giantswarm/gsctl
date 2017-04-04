package util

import "testing"

type TruncateTest struct {
	input  string
	length int
	output string
}

type KeypairIDTest struct {
	input  string
	output string
}

var truncateTests = []TruncateTest{
	{"This is a longer string", 10, "This is a…"},
	{"Cut after a blank character", 19, "Cut after a blank …"},
	{"This won't be touched", 30, "This won't be touched"},
}

var keypairIDTests = []KeypairIDTest{
	{"a1:b2:c3:d4:e5:f6:g7:00", "a1b2c3d4e5f6g700"},
}

func TestTruncate(t *testing.T) {
	for i, test := range truncateTests {
		out := Truncate(test.input, test.length)
		if out != test.output {
			t.Errorf("#%d: TestTruncate(%s, %d) = '%s'; want '%s'", i, test.input, test.length, out, test.output)
		}
	}
}

func TestCleanKeypairID(t *testing.T) {
	for i, test := range keypairIDTests {
		out := CleanKeypairID(test.input)
		if out != test.output {
			t.Errorf("#%d: CleanKeypairID(%s) = '%s'; want '%s'", i, test.input, out, test.output)
		}
	}
}
