package table

import (
	"github.com/Masterminds/semver"

	"github.com/giantswarm/gsctl/util"
)

var SortableTypes = struct {
	String string
	Semver string
	Date   string
}{
	String: "string",
	Semver: "semver",
	Date:   "date",
}

var SortableDirections = struct {
	ASC  string
	DESC string
}{
	ASC:  "asc",
	DESC: "desc",
}

type Sortable struct {
	SortType string
}

func CompareStrings(a string, b string, direction string) bool {
	if direction == SortableDirections.DESC {
		return a > b
	}

	return a < b
}

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
	if direction == SortableDirections.DESC {
		return cmp > 0
	}

	return cmp <= 0
}

func CompareDates(a string, b string, direction string) bool {
	dateA := util.ParseDate(a)
	dateB := util.ParseDate(b)

	cmp := dateA.After(dateB)
	if direction == SortableDirections.DESC {
		return cmp
	}

	return cmp == false
}

func GetCompareFunc(t string) func(string, string, string) bool {
	switch t {
	case SortableTypes.String:
		return CompareStrings

	case SortableTypes.Date:
		return CompareDates

	case SortableTypes.Semver:
		return CompareSemvers

	default:
		return CompareStrings
	}
}

type CompareFunc func(int, int) bool
