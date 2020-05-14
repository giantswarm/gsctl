package table

import (
	"sort"
	"strings"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

type Table struct {
	columns []Column
	rows    [][]string

	columnizeConfig *columnize.Config
}

func New() Table {
	t := Table{}

	t.columnizeConfig = columnize.DefaultConfig()
	t.columnizeConfig.Glue = "   "

	return t
}

func (t *Table) SetColumns(c []Column) {
	t.columns = c[:]
}

func (t *Table) SetRows(r [][]string) {
	t.rows = r[:][:]
}

func (t *Table) String() string {
	rows := make([]string, 0, len(t.rows)+1)

	{
		columns := make([]string, 0, len(t.columns))
		for _, col := range t.columns {
			columns = append(columns, col.GetHeader())
		}
		rows = append(rows, strings.Join(columns, "|"))
	}

	{
		for _, row := range t.rows {
			rows = append(rows, strings.Join(row, "|"))
		}
	}

	formattedTable := columnize.Format(rows, t.columnizeConfig)

	return formattedTable
}

func (t *Table) SortByColumnName(n string, direction string) error {
	var (
		colIndex int
		column   Column
	)
	{
		for i, col := range t.columns {
			if col.Name == n {
				colIndex = i
				column = col

				break
			}
		}
		if column.Name == "" {
			return microerror.Mask(columnNotFoundError)
		}
	}

	sortDir := direction
	if sortDir != sortable.Directions.ASC && sortDir != sortable.Directions.DESC {
		sortDir = sortable.Directions.ASC
	}

	compareFunc := sortable.GetCompareFunc(column.SortType)
	sort.Slice(t.rows, func(i, j int) bool {
		return compareFunc(t.rows[i][colIndex], t.rows[j][colIndex], direction)
	})

	return nil
}
