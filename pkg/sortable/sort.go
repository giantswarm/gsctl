package sortable

import (
	"github.com/Masterminds/semver"

	"github.com/giantswarm/gsctl/util"
)

// Types represent the kinds of data that can be in a string field.
var Types = struct {
	String string
	Semver string
	Date   string
}{
	String: "string",
	Semver: "semver",
	Date:   "date",
}

// Directions represent the sorting direction possibilities.
var Directions = struct {
	ASC  string
	DESC string
}{
	ASC:  "asc",
	DESC: "desc",
}

// Sortable makes the data type that embeds it, sortable.
type Sortable struct {
	SortType string
}

// CompareStrings represents the comparison algorithm for regular strings.
func CompareStrings(a string, b string, direction string) bool {
	if direction == Directions.DESC {
		return a > b
	}

	return a < b
}

// CompareSemvers represents the comparison algorithm for string-encoded versions, in semver format.
func CompareSemvers(a string, b string, direction string) bool {
	verA, err := semver.NewVersion(a)
	if err != nil {
		return false
	}

	verB, err := semver.NewVersion(b)
	if err != nil {
		return false
	}

	cmp := verA.Compare(verB)
	if direction == Directions.DESC {
		return cmp > 0
	}

	return cmp <= 0
}

// CompareDates represents the comparison algorithm for string-encoded dates.
func CompareDates(a string, b string, direction string) bool {
	dateA := util.ParseShortDate(a)
	dateB := util.ParseShortDate(b)

	cmp := dateA.After(dateB)
	if direction == Directions.DESC {
		return cmp
	}

	return cmp == false
}

// GetCompareFunc gets the right comparison algorithm for the provided type.
func GetCompareFunc(t string) func(string, string, string) bool {
	switch t {
	case Types.String:
		return CompareStrings

	case Types.Date:
		return CompareDates

	case Types.Semver:
		return CompareSemvers

	default:
		return CompareStrings
	}
}
