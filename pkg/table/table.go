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

func (t *Table) SortByColumnName(n string, direction string) error {
	if len(t.rows) < 2 {
		return nil
	}

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
			return microerror.Mask(fieldNotFoundError)
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

func (t *Table) GetColumnNameFromInitials(i string) (string, error) {
	i = strings.ToLower(i)

	var (
		columnNames   = make([]string, 0, len(t.columns))
		matchingNames []string
	)
	{
		for _, col := range t.columns {
			if col.Name != "" {
				columnNames = append(columnNames, col.Name)

				if strings.HasPrefix(strings.ToLower(col.Name), i) {
					matchingNames = append(matchingNames, col.Name)
				}
			}
		}
	}

	if len(matchingNames) == 0 {
		return "", microerror.Mask(fieldNotFoundError)
	} else if len(matchingNames) > 1 {
		return "", microerror.Mask(multipleFieldsMatchingError)
	}

	return matchingNames[0], nil
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
