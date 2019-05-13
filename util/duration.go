package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	// roughly the largest time.Duration values representable
	maxDurationInHours  = 2562047
	maxDurationInDays   = 106751
	maxDurationInWeeks  = 15250
	maxDurationInMonths = 3558
	maxDurationInYears  = 292
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

// ParseDuration converts strings like "1d" into a duration. Only
// one combination of <number> and <unit> is allowed, and the unit
// must be one of:
// - "h" - one hour
// - "d" - day (24 hours)
// - "w" - 7 days
// - "m" - 30 days
// - "y" - 365 days
//
// This is necessary because time.ParseDuration does not support units
// larger than hour.
func ParseDuration(durationString string) (time.Duration, error) {
	var duration time.Duration

	pattern := regexp.MustCompile(`^([0-9]+)([hdwmy])$`)

	match := pattern.FindStringSubmatch(durationString)

	if len(match) != 3 {
		return duration, microerror.Mask(InvalidDurationStringError)
	}

	numberInt, err := strconv.Atoi(match[1])
	if err != nil {
		return duration, microerror.Mask(InvalidDurationStringError)
	}

	number := int64(numberInt)

	unit := match[2]

	//

	switch unit {
	case "h":
		if number > maxDurationInHours {
			return duration, microerror.Mask(DurationExceededError)
		}
		duration = 3600 * time.Duration(number) * time.Second
	case "d":
		if number > maxDurationInDays {
			return duration, microerror.Mask(DurationExceededError)
		}
		duration = 24 * 3600 * time.Duration(number) * time.Second
	case "w":
		if number > maxDurationInWeeks {
			return duration, microerror.Mask(DurationExceededError)
		}
		duration = 7 * 24 * 3600 * time.Duration(number) * time.Second
	case "m":
		if number > maxDurationInMonths {
			return duration, microerror.Mask(DurationExceededError)
		}
		duration = 30 * 24 * 3600 * time.Duration(number) * time.Second
	case "y":
		if number > maxDurationInYears {
			return duration, microerror.Mask(DurationExceededError)
		}
		duration = 365 * 24 * 3600 * time.Duration(number) * time.Second
	default:
		return duration, microerror.Mask(InvalidDurationStringError)
	}

	return duration, nil
}
