package util

import (
	"testing"
	"time"
)

var friendlyDurationTest = []struct {
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
	for _, tt := range friendlyDurationTest {
		phrase := DurationPhrase(tt.in)
		if phrase != tt.out {
			t.Errorf("Value '%d', got '%s', wanted '%s'", tt.in, phrase, tt.out)
		}
	}
}

var parseDurationTest = []struct {
	in  string
	out time.Duration
}{
	{"1h", 1 * time.Hour},
	{"1d", 24 * time.Hour},
	{"1w", 7 * 24 * time.Hour},
	{"1m", 30 * 24 * time.Hour},
	{"1y", 365 * 24 * time.Hour},
}

func TestParseDuration(t *testing.T) {
	for _, tt := range parseDurationTest {
		duration, err := ParseDuration(tt.in)
		if err != nil {
			t.Errorf("Value '%s' yielded error: '%s'", tt.in, err)
		} else if duration != tt.out {
			t.Errorf("Value '%s', got '%v', wanted '%v'", tt.in, duration, tt.out)
		}
	}
}
