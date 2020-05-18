package table

import (
	"sort"
	"strings"

	"github.com/giantswarm/columnize"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

// Table represents a data structure that can hold and display the contents of a table.
type Table struct {
	columns []Column
	rows    [][]string

	// columnizeConfig represents the configuration of the table formatter.
	columnizeConfig *columnize.Config
}

// New creates a new Table.
func New() Table {
	t := Table{}

	t.columnizeConfig = columnize.DefaultConfig()
	t.columnizeConfig.Glue = "   "

	return t
}

// SetColumns sets the table's columns.
func (t *Table) SetColumns(c []Column) {
	t.columns = c[:]
}

// SetRows sets the table's rows.
func (t *Table) SetRows(r [][]string) {
	t.rows = r[:][:]
}

// SortByColumnName sorts the table by a column name, in the given direction.
func (t *Table) SortByColumnName(n string, direction string) error {
	// Skip if there is nothing to sort, or if there's no column name provided.
	if len(t.rows) < 2 || n == "" {
		return nil
	}

	colIndex, column, err := t.GetColumnByName(n)
	if err != nil {
		return microerror.Mask(err)
	}

	// Default to Ascending direction sorting.
	sortDir := direction
	if sortDir != sortable.ASC && sortDir != sortable.DESC {
		sortDir = sortable.ASC
	}

	// Get the comparison algorithm for the current sorting type.
	compareFunc := sortable.GetCompareFunc(column.SortType)
	sort.Slice(t.rows, func(i, j int) bool {
		var iVal string
		{
			if colIndex >= len(t.rows[i]) {
				iVal = "n/a"
			} else {
				iVal = RemoveColors(t.rows[i][colIndex])
			}
		}

		var jVal string
		{
			if colIndex >= len(t.rows[j]) {
				jVal = "n/a"
			} else {
				jVal = RemoveColors(t.rows[j][colIndex])
			}
		}

		return compareFunc(iVal, jVal, direction)
	})

	return nil
}

// GetColumnByName fetches the index and data structure of a column, by knowing its name.
func (t *Table) GetColumnByName(n string) (int, Column, error) {
	var (
		colIndex int
		column   Column
	)
	for i, col := range t.columns {
		if col.Name == n {
			colIndex = i
			column = col

			break
		}
	}
	if column.Name == "" {
		return 0, Column{}, microerror.Mask(fieldNotFoundError)
	}

	return colIndex, column, nil
}

// GetColumnNameFromInitials matches a given input with a name of an existent column,
// without caring about casing, or if the given input is the complete name of the column.
func (t *Table) GetColumnNameFromInitials(i string) (string, error) {
	i = strings.ToLower(i)

	var (
		columnNames   = make([]string, 0, len(t.columns))
		matchingNames []string
	)
	for _, col := range t.columns {
		if col.Name != "" {
			columnNames = append(columnNames, col.Name)

			nameLowerCased := strings.ToLower(col.Name)
			if strings.HasPrefix(nameLowerCased, i) {
				matchingNames = append(matchingNames, col.Name)

				if nameLowerCased == i {
					return matchingNames[0], nil
				}
			}
		}
	}

	if len(matchingNames) == 0 {
		return "", microerror.Maskf(fieldNotFoundError, "available fields for sorting: %v", strings.Join(columnNames, ", "))
	} else if len(matchingNames) > 1 {
		return "", microerror.Maskf(multipleFieldsMatchingError, "%v", strings.Join(matchingNames, ", "))
	}

	return matchingNames[0], nil
}

// String makes the Table data structure implement the Stringer interface,
// so we can easily pretty-print it.
func (t *Table) String() string {
	rows := make([]string, 0, len(t.rows)+1)

	{
		columns := make([]string, 0, len(t.columns))
		for _, col := range t.columns {
			if !col.Hidden {
				columns = append(columns, col.GetHeader())
			}
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
