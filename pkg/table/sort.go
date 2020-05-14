package table

import (
	"sort"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

func SortMapSliceUsingColumnData(mapSlice []map[string]interface{}, byCol Column, fieldMapping map[string]string) {
	compareFunc := sortable.GetCompareFunc(byCol.SortType)
	sort.Slice(mapSlice, func(i, j int) bool {
		iField := "n/a"
		{
			iValue, ok := mapSlice[i][fieldMapping[byCol.Name]]
			if ok {
				iField = iValue.(string)
			}
		}

		jField := "n/a"
		{
			jValue, ok := mapSlice[j][fieldMapping[byCol.Name]]
			if ok {
				jField = jValue.(string)
			}
		}

		return compareFunc(iField, jField, sortable.Directions.ASC)
	})
}
