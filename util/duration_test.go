package util

import "testing"

var durationTest = []struct {
	in  int
	out string
}{
	{1, "1 hour"},
	{12, "12 hours"},
	{24, "1 day"},
	{25, "1 day, 1 hour"},
	{26, "1 day, 2 hours"},
	{48, "2 days"},
	{49, "2 days, 1 hour"},
	{75, "3 days, 3 hours"},
	{24 * 7, "1 week"},
	{24*7*2 - 5, "1 week, 6 days"},
	{24 * 7 * 2, "2 weeks"},
	{24*7*2 + 5, "2 weeks, 5 hours"},
	{11 * 24 * 30, "11 months"},
	{365 * 24, "1 year"},
	{365*24 + 4, "1 year, 4 hours"},
	{365*24 + 30, "1 year, 1 day"},
	{365*24*2 + 30, "2 years, 1 day"},
}

func TestFriendlyDuration(t *testing.T) {
	for _, tt := range durationTest {
		phrase := DurationPhrase(tt.in)
		if phrase != tt.out {
			t.Errorf("Value '%d', got '%s', wanted '%s'", tt.in, phrase, tt.out)
		}
	}
}
