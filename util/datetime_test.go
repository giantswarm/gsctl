package util

import (
	"testing"
	"time"
)

type Test struct {
	input       string
	timeValue   time.Time
	outputShort string
}

var tests = []Test{
	{
		"2006-01-02T15:04:05.000Z",
		time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		"2006 Jan 02, 15:04"},
	{
		"1999-11-24T00:57:28.999999Z",
		time.Date(1999, time.November, 24, 0, 57, 28, 0, time.UTC),
		"1999 Nov 24, 00:57",
	},
}

func TestParseDate(t *testing.T) {
	for i, test := range tests {
		out := ParseDate(test.input)
		if out != test.timeValue {
			t.Errorf("#%d: ParseDate(%s)=%s; want %s", i, test.input, out, test.timeValue)
		}
	}
}

func TestShortDate(t *testing.T) {
	for i, test := range tests {
		out := ShortDate(test.timeValue)
		if out != test.outputShort {
			t.Errorf("#%d: ShortDate(%+v)='%s'; want '%s'", i, test.timeValue, out, test.outputShort)
		}
	}
}
