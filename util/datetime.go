package util

import (
	"regexp"
	"time"
)

const (
	shortFormat string = "2006 Jan 02, 15:04 UTC"
)

// ParseDate parses our common date/time strings
// and returns a time.Time object
func ParseDate(dateString string) time.Time {
	// our standard format
	template := "2006-01-02T15:04:05.000Z"

	// normalizing the number of decimal places to 3
	re := regexp.MustCompile("\\.[0-9]+Z$")
	dateString = re.ReplaceAllLiteralString(dateString, ".000Z")

	t, _ := time.Parse(template, dateString)
	return t
}

// ShortDate reformats a time.Time to a short date
func ShortDate(date time.Time) string {
	return date.Format(shortFormat)
}
