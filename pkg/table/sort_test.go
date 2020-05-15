package table

import (
	"strconv"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

func Test_SortMapSliceUsingColumnData(t *testing.T) {
	testCases := []struct {
		mapSlice       []map[string]interface{}
		column         Column
		fieldMapping   map[string]string
		expectedResult []map[string]interface{}
	}{
		{
			mapSlice: []map[string]interface{}{
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
			},
			column: Column{
				Name: "name",
				Sortable: sortable.Sortable{
					SortType: sortable.Types.String,
				},
			},
			fieldMapping: map[string]string{
				"id":      "id",
				"name":    "name",
				"created": "creation_date",
				"release": "release_version",
			},
			expectedResult: []map[string]interface{}{
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
			},
		},
		{
			mapSlice: []map[string]interface{}{
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
			},
			column: Column{
				Name: "created",
				Sortable: sortable.Sortable{
					SortType: sortable.Types.Date,
				},
			},
			fieldMapping: map[string]string{
				"id":      "id",
				"name":    "name",
				"created": "creation_date",
				"release": "release_version",
			},
			expectedResult: []map[string]interface{}{
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
			},
		},
		{
			mapSlice: []map[string]interface{}{
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
			},
			column: Column{
				Name: "release",
				Sortable: sortable.Sortable{
					SortType: sortable.Types.Date,
				},
			},
			fieldMapping: map[string]string{
				"id":      "id",
				"name":    "name",
				"created": "creation_date",
				"release": "release_version",
			},
			expectedResult: []map[string]interface{}{
				{
					"id":              "saf91",
					"name":            "Some very different name",
					"creation_date":   "2020-01-03T15:21:05.000Z",
					"release_version": "9.0.1",
				},
				{
					"id":              "as712",
					"name":            "Some name",
					"creation_date":   "2020-01-02T15:04:05.000Z",
					"release_version": "12.0.1",
				},
				{
					"id":              "d91ns",
					"name":            "Some other name",
					"creation_date":   "2020-01-03T15:04:05.000Z",
					"release_version": "12.0.1",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			SortMapSliceUsingColumnData(tc.mapSlice, tc.column, tc.fieldMapping)

			if diff := cmp.Diff(tc.expectedResult, tc.mapSlice); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}
		})
	}
}

func Test_RemoveColors(t *testing.T) {
	testCases := []struct {
		input          string
		expectedResult string
	}{
		{
			input:          color.CyanString("some-string"),
			expectedResult: "some-string",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := RemoveColors(tc.input)

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %s, got %s", i, tc.expectedResult, result)
			}
		})
	}
}
