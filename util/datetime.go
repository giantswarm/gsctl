package util

import (
	"regexp"
	"strconv"
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
	re1 := regexp.MustCompile(":([0-9]{2})Z")
	re2 := regexp.MustCompile("\\.[0-9]+")
	dateString = re1.ReplaceAllString(dateString, ":$1.000Z")
	dateString = re2.ReplaceAllLiteralString(dateString, ".000")

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

	t = parseShortDate(dateString)

	return t
}

// ShortDate reformats a time.Time to a short date
func ShortDate(date time.Time) string {
	return date.Format(shortFormat)
}

func parseShortDate(dateStr string) time.Time {
	// 'YYYY MMM DD HH:MM LOC'
	if len(dateStr) < 18 {
		return time.Time{}
	}

	year, _ := strconv.Atoi(dateStr[:4])
	month, _ := time.Parse("Jan", dateStr[5:8])
	days, _ := strconv.Atoi(dateStr[9:11])
	hours, _ := strconv.Atoi(dateStr[13:15])
	minutes, _ := strconv.Atoi(dateStr[16:18])

	newDate := time.Date(year, month.Month(), days, hours, minutes, 0, 0, time.UTC)

	return newDate
}
