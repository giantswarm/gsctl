package util

import (
	"regexp"
	"time"
)

const (
	shortFormat = "2006 Jan 02, 15:04 UTC"
)

// ParseDate parses our common date/time strings
// and returns a time.Time object
func ParseDate(dateString string) time.Time {
	// our standard format
	template := "2006-01-02T15:04:05.000Z"
	alternativeTemplate := "2006-01-02T15:04:05.000-07:00"

	// normalizing the number of decimal places to 3
	// and discarding sub-second detail along the way
	re := regexp.MustCompile("\\.[0-9]+")
	dateString = re.ReplaceAllLiteralString(dateString, ".000")

	// try parsing with several formats
	t, err := time.Parse(template, dateString)
	if err == nil {
		return t
	}

	t, err = time.Parse(alternativeTemplate, dateString)
	if err == nil {
		l, _ := time.LoadLocation("UTC")
		t = t.In(l)
	}
	return t
}

// ShortDate reformats a time.Time to a short date
func ShortDate(date time.Time) string {
	return date.Format(shortFormat)
}
