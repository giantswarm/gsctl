package formatting

import (
	"sort"
	"strings"
)

// AvailabilityZonesList returns the list of availability zones
// as one string consisting of uppercase letters only, separated by comma.
// Example: "A,B,C".
func AvailabilityZonesList(az []string) string {
	var shortened []string

	for _, zone := range az {
		lastLetterIndex := len(zone) - 1
		if lastLetterIndex < 0 {
			lastLetterIndex = 0
		}

		// last character of each item
		shortened = append(shortened, zone[lastLetterIndex:])
	}

	sort.Strings(shortened)

	return strings.ToUpper(strings.Join(shortened, ","))
}
