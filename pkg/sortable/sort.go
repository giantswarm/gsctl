package sortable

import (
	"strings"

	"github.com/Masterminds/semver"

	stringsort "github.com/facette/natsort"

	"github.com/giantswarm/gsctl/util"
)

// The kinds of data that can be in a string field.
const (
	String = "string"
	Semver = "semver"
	Date   = "date"
)

// The sorting direction possibilities.
const (
	ASC  = "asc"
	DESC = "desc"
)

// Sortable makes the data type that embeds it, sortable.
type Sortable struct {
	SortType string
}

// CompareStrings represents the comparison algorithm for regular strings.
func CompareStrings(a string, b string, direction string) bool {
	result := stringsort.Compare(strings.ToLower(a), strings.ToLower(b))

	if direction == DESC {
		return !result
	}

	return result
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
	if direction == DESC {
		return cmp > 0
	}

	return cmp <= 0
}

// CompareDates represents the comparison algorithm for string-encoded dates.
func CompareDates(a string, b string, direction string) bool {
	dateA := util.ParseDate(a)
	dateB := util.ParseDate(b)

	cmp := dateA.After(dateB)
	if direction == DESC {
		return cmp
	}

	return cmp == false
}

// GetCompareFunc gets the right comparison algorithm for the provided type.
func GetCompareFunc(t string) func(string, string, string) bool {
	switch t {
	case String:
		return CompareStrings

	case Date:
		return CompareDates

	case Semver:
		return CompareSemvers

	default:
		return CompareStrings
	}
}
