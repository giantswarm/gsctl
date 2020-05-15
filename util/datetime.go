package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
	newDate := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

	dateStr = strings.Replace(dateStr, ",", "", -1)
	dateParts := strings.Split(dateStr, " ")

	// [ YYYY MMM DD HH:MM UTC]
	if len(dateParts) < 5 {
		return newDate
	}

	// Only transform the last 4 digits of the year to int,
	// as sometimes (if the string content is colored)
	// there may be some random utf8 chars before.
	yearDigitsToTransform := len(dateParts[0]) - 4
	year, _ := strconv.Atoi(dateParts[0][yearDigitsToTransform:])
	month, _ := time.Parse("Jan", dateParts[1])
	days, _ := strconv.Atoi(dateParts[2])
	newDate = newDate.AddDate(year, int(month.Month()), days)

	timeDigits := strings.Split(dateParts[3], ":")
	timeDuration, _ := time.ParseDuration(fmt.Sprintf("%sh%sm", timeDigits[0], timeDigits[1]))
	newDate = newDate.Add(timeDuration)

	return newDate
}
