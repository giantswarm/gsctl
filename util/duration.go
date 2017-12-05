package util

import (
	"fmt"
	"strings"
)

// DurationPhrase creates a human-friendly phrase from a number of hours
// expressing a duration, like "3 days, 2 hours". Precision of output
// is limited in favour of readability.
func DurationPhrase(hours int) string {

	hoursInYear := 365 * 24
	hoursInMonth := 30 * 24
	hoursInWeek := 7 * 24
	hoursInDay := 24

	years := hours / hoursInYear
	months := (hours - years*hoursInYear) / hoursInMonth
	weeks := (hours - years*hoursInYear - months*hoursInMonth) / hoursInWeek
	days := (hours - years*hoursInYear - months*hoursInMonth - weeks*hoursInWeek) / hoursInDay
	hours = hours - years*hoursInYear - months*hoursInMonth - weeks*hoursInWeek - days*hoursInDay

	phraseParts := []string{}

	if years > 0 {
		plural := ""
		if years > 1 {
			plural = "s"
		}
		phraseParts = append(phraseParts, fmt.Sprintf("%d year%s", years, plural))
	}

	if months > 0 {
		plural := ""
		if months > 1 {
			plural = "s"
		}
		phraseParts = append(phraseParts, fmt.Sprintf("%d month%s", months, plural))
	}

	if weeks > 0 && len(phraseParts) < 2 {
		plural := ""
		if weeks > 1 {
			plural = "s"
		}
		phraseParts = append(phraseParts, fmt.Sprintf("%d week%s", weeks, plural))
	}

	if days > 0 && len(phraseParts) < 2 {
		plural := ""
		if days > 1 {
			plural = "s"
		}
		phraseParts = append(phraseParts, fmt.Sprintf("%d day%s", days, plural))
	}

	if hours > 0 && len(phraseParts) < 2 {
		plural := ""
		if hours > 1 {
			plural = "s"
		}
		phraseParts = append(phraseParts, fmt.Sprintf("%d hour%s", hours, plural))
	}

	return strings.Join(phraseParts, ", ")
}
