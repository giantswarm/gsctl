package formatting

import (
	"sort"
	"strings"
)

// AvailabilityZonesList returns the list of availability zones
// as one string consisting of uppercase letters only, separated by comma.
// Example: "A,B,C".
func AvailabilityZonesList(az []string) string {
	shortened := []string{}

	for _, az := range az {
		// last character of each item
		shortened = append(shortened, az[len(az)-1:])
	}

	sort.Strings(shortened)

	return strings.ToUpper(strings.Join(shortened, ","))
}
