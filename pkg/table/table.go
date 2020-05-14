package table

import (
	"strings"

	"github.com/giantswarm/columnize"
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
