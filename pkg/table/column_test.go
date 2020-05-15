package table

import (
	"strconv"
	"testing"
)

func Test_GetHeader(t *testing.T) {
	testCases := []struct {
		column         Column
		expectedResult string
	}{
		{
			column: Column{
				Name: "some-name",
			},
			expectedResult: "some-name",
		},
		{
			column: Column{
				Name:        "some-name",
				DisplayName: "Cool Display Name",
			},
			expectedResult: "Cool Display Name",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := tc.column.GetHeader()

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %s, got %s", i, tc.expectedResult, result)
			}
		})
	}
}
