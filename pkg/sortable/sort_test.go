package sortable

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_GetCompareFunc(t *testing.T) {
	testCases := []struct {
		sortableType string
		fn           func(string, string, string) bool
	}{
		{
			sortableType: Types.String,
			fn:           CompareStrings,
		}, {
			sortableType: Types.Date,
			fn:           CompareDates,
		},
		{
			sortableType: Types.Semver,
			fn:           CompareSemvers,
		},
		{
			sortableType: "random",
			fn:           CompareStrings,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := reflect.ValueOf(GetCompareFunc(tc.sortableType))
			expectedFn := reflect.ValueOf(tc.fn)

			if result.Pointer() != expectedFn.Pointer() {
				t.Errorf("Case %d - Result did not match", i)
			}
		})
	}
}

func Test_CompareStrings(t *testing.T) {
	testCases := []struct {
		a              string
		b              string
		direction      string
		expectedResult bool
	}{
		{
			a:              "some-string",
			b:              "some-other-string",
			direction:      Directions.ASC,
			expectedResult: false,
		},
		{
			a:              "some-string",
			b:              "some-other-string",
			direction:      Directions.DESC,
			expectedResult: true,
		},
		{
			a:              "12312assdaads",
			b:              "some-string",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "12312assdaads",
			b:              "some-string",
			direction:      Directions.DESC,
			expectedResult: false,
		},
		{
			a:              "_!ldsanl",
			b:              "some-string",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "_!ldsanl",
			b:              "some-string",
			direction:      Directions.DESC,
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := CompareStrings(tc.a, tc.b, tc.direction)

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %t, got %t", i, tc.expectedResult, result)
			}
		})
	}
}

func Test_CompareSemvers(t *testing.T) {
	testCases := []struct {
		a              string
		b              string
		direction      string
		expectedResult bool
	}{
		{
			a:              "1.0.0",
			b:              "1.0.1",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "1.0.0",
			b:              "1.0.1",
			direction:      Directions.DESC,
			expectedResult: false,
		},
		{
			a:              "0.0.9",
			b:              "1.0.1",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "0.0.9",
			b:              "1.0.1",
			direction:      Directions.DESC,
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := CompareSemvers(tc.a, tc.b, tc.direction)

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %t, got %t", i, tc.expectedResult, result)
			}
		})
	}
}

func Test_CompareDates(t *testing.T) {
	testCases := []struct {
		a              string
		b              string
		direction      string
		expectedResult bool
	}{
		{
			a:              "1999 Nov 24, 00:57 UTC",
			b:              "2016 Dec 05, 14:41 UTC",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "1999 Nov 24, 00:57 UTC",
			b:              "2016 Dec 05, 14:41 UTC",
			direction:      Directions.DESC,
			expectedResult: false,
		},
		{
			a:              "1999-11-24T00:57:28.999999Z",
			b:              "2006-01-02T15:04:05.000Z",
			direction:      Directions.ASC,
			expectedResult: true,
		},
		{
			a:              "1999-11-24T00:57:28.999999Z",
			b:              "2006-01-02T15:04:05.000Z",
			direction:      Directions.DESC,
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := CompareDates(tc.a, tc.b, tc.direction)

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %t, got %t", i, tc.expectedResult, result)
			}
		})
	}
}
