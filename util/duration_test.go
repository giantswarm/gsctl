package util

import (
	"testing"
	"time"
)

// TestFriendlyDuration tests the conversion from duration integers
// into user-friendly phrases.
func TestFriendlyDuration(t *testing.T) {
	var testCases = []struct {
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

	for _, tc := range testCases {
		phrase := DurationPhrase(tc.in)
		if phrase != tc.out {
			t.Errorf("Value '%d', got '%s', wanted '%s'", tc.in, phrase, tc.out)
		}
	}
}

// TestParseDuration tests the parsing of durations into time.Duration values
func TestParseDuration(t *testing.T) {
	var testCases = []struct {
		in  string
		out time.Duration
	}{
		{"1h", 1 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"1m", 30 * 24 * time.Hour},
		{"1y", 365 * 24 * time.Hour},
		{"10y", 10 * 365 * 24 * time.Hour},
		{"100y", 100 * 365 * 24 * time.Hour},
		{"2562047h", 2562047 * time.Hour},
		{"106751d", 106751 * 24 * time.Hour},
		{"15250w", 15250 * 7 * 24 * time.Hour},
		{"3558m", 3558 * 30 * 24 * time.Hour},
		{"292y", 292 * 365 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		duration, err := ParseDuration(tc.in)
		if err != nil {
			t.Errorf("Value '%s' yielded error: '%s'", tc.in, err)
		} else if duration != tc.out {
			t.Errorf("Value '%s', got '%v', wanted '%v'", tc.in, duration, tc.out)
		}
	}
}

// TestParseDurationError tests the parsing of durations which is supposed to fail
func TestParseDurationError(t *testing.T) {
	testCases := []struct {
		in           string
		errorMatcher func(error) bool
	}{
		{"", IsInvalidDurationStringError},
		{"10f", IsInvalidDurationStringError},
		{"293y", IsDurationExceededError},
		{"3559m", IsDurationExceededError},
		{"15251w", IsDurationExceededError},
		{"106752d", IsDurationExceededError},
		{"2562048h", IsDurationExceededError},
	}

	for _, tc := range testCases {
		_, err := ParseDuration(tc.in)
		if !tc.errorMatcher(err) {
			t.Errorf("test case '%s': Expected error, got '%#v'", tc.in, err)
		}
	}
}
