package table

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

func Test_SetColumns(t *testing.T) {
	testCases := []struct {
		columns        []Column
		expectedResult []Column
	}{
		{
			columns: []Column{
				{
					Name: "some-col",
				},
				{
					Name: "some-other-col",
				},
				{
					Name: "some-random-col",
				},
			},
			expectedResult: []Column{
				{
					Name: "some-col",
				},
				{
					Name: "some-other-col",
				},
				{
					Name: "some-random-col",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetColumns(tc.columns)

			if diff := cmp.Diff(tc.expectedResult, table.columns); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}
		})
	}
}

func Test_SetRows(t *testing.T) {
	testCases := []struct {
		rows           [][]string
		expectedResult [][]string
	}{
		{
			rows: [][]string{
				{
					"something",
					"some other thing",
					"a third thing",
				},
				{
					"something",
					"some other thing",
					"a third thing",
				},
			},
			expectedResult: [][]string{
				{
					"something",
					"some other thing",
					"a third thing",
				},
				{
					"something",
					"some other thing",
					"a third thing",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetRows(tc.rows)

			if diff := cmp.Diff(tc.expectedResult, table.rows); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}
		})
	}
}

func Test_SortByColumnName(t *testing.T) {
	testCases := []struct {
		columns        []Column
		rows           [][]string
		sortBy         string
		direction      string
		expectedResult [][]string
		errorMatcher   func(error) bool
	}{
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "some-col",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "some-col",
			direction: "",
			expectedResult: [][]string{
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "some-col",
			direction: sortable.DESC,
			expectedResult: [][]string{
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "some-other-col",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "some-random-col",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			sortBy:    "naaah",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
			},
			errorMatcher: IsFieldNotFoundError,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
			},
			sortBy:    "some-other-col",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
					Hidden: true,
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
				},
			},
			sortBy:    "some-random-col",
			direction: sortable.ASC,
			expectedResult: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
				},
			},
			errorMatcher: nil,
		},
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
					Sortable: sortable.Sortable{
						SortType: sortable.String,
					},
				},
				{
					Name: "some-other-col",
					Sortable: sortable.Sortable{
						SortType: sortable.Date,
					},
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
					Sortable: sortable.Sortable{
						SortType: sortable.Semver,
					},
					Hidden: true,
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"12.0.1",
				},
			},
			sortBy:    "some-random-col",
			direction: sortable.DESC,
			expectedResult: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"12.0.1",
				},
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetColumns(tc.columns)
			table.SetRows(tc.rows)

			err := table.SortByColumnName(tc.sortBy, tc.direction)

			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Unexpected error: %s", i, err)
				}
			} else {
				if diff := cmp.Diff(tc.expectedResult, tc.rows); diff != "" {
					t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
				}
			}
		})
	}
}

func Test_GetColumnByName(t *testing.T) {
	testCases := []struct {
		columns             []Column
		name                string
		expectedResult      Column
		expectedResultIndex int
		errorMatcher        func(error) bool
	}{
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "date",
				},
				{
					Name: "release",
				},
			},
			name:                "name",
			expectedResult:      Column{Name: "name"},
			expectedResultIndex: 0,
			errorMatcher:        nil,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "date",
				},
				{
					Name: "release",
				},
			},
			name:                "game",
			expectedResult:      Column{},
			expectedResultIndex: 0,
			errorMatcher:        IsFieldNotFoundError,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "date",
				},
				{
					Name: "release",
				},
			},
			name:                "date",
			expectedResult:      Column{Name: "date"},
			expectedResultIndex: 1,
			errorMatcher:        nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetColumns(tc.columns)
			index, result, err := table.GetColumnByName(tc.name)

			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Unexpected error: %s", i, err)
				}
			} else {
				if index != tc.expectedResultIndex {
					t.Errorf("Case %d - Result index did not match.\nOutput: %d", i, index)
				}

				if diff := cmp.Diff(tc.expectedResult, result); diff != "" {
					t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
				}
			}
		})
	}
}

func Test_GetColumnNameFromInitials(t *testing.T) {
	testCases := []struct {
		columns        []Column
		initials       string
		expectedResult string
		errorMatcher   func(error) bool
	}{
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
			},
			initials:       "name",
			expectedResult: "name",
			errorMatcher:   nil,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
			},
			initials:       "n",
			expectedResult: "name",
			errorMatcher:   nil,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
			},
			initials:       "NaM",
			expectedResult: "name",
			errorMatcher:   nil,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
			},
			initials:       "d",
			expectedResult: "dAtE",
			errorMatcher:   nil,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
			},
			initials:       "flower",
			expectedResult: "",
			errorMatcher:   IsFieldNotFoundError,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
				{
					Name: "religion",
				},
			},
			initials:       "r",
			expectedResult: "",
			errorMatcher:   IsMultipleFieldsMatchingError,
		},
		{
			columns: []Column{
				{
					Name: "name",
				},
				{
					Name: "dAtE",
				},
				{
					Name: "Release",
				},
				{
					Name: "release-date",
				},
			},
			initials:       "release",
			expectedResult: "Release",
			errorMatcher:   nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetColumns(tc.columns)
			result, err := table.GetColumnNameFromInitials(tc.initials)

			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Errorf("Case %d - Unexpected error: %s", i, err)
				}
			} else if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %s, got %s", i, tc.expectedResult, result)
			}
		})
	}
}

func Test_String(t *testing.T) {
	testCases := []struct {
		columns        []Column
		rows           [][]string
		expectedResult string
	}{
		{
			columns: []Column{
				{
					Name:        "some-col",
					DisplayName: "SOME COLUMN",
				},
				{
					Name: "some-other-col",
				},
				{
					Name:        "some-random-col",
					DisplayName: "Some Random Column",
				},
			},
			rows: [][]string{
				{
					"Good dog",
					"2016 Dec 05, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good cat",
					"2016 Dec 25, 14:41 UTC",
					"12.0.1",
				},
				{
					"Good parrot",
					"2016 Dec 25, 15:41 UTC",
					"9.0.1",
				},
			},
			expectedResult: `SOME COLUMN   some-other-col           Some Random Column
Good dog      2016 Dec 05, 14:41 UTC   12.0.1
Good cat      2016 Dec 25, 14:41 UTC   12.0.1
Good parrot   2016 Dec 25, 15:41 UTC   9.0.1`,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			table := New()
			table.SetColumns(tc.columns)
			table.SetRows(tc.rows)

			result := table.String()

			if diff := cmp.Diff(tc.expectedResult, result); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}
		})
	}
}
