package table

import (
	"github.com/fatih/color"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

// Column represents the data structure of a table column.
type Column struct {
	sortable.Sortable
	// Name represents the column name that will be used for sorting,
	// and as a default table header.
	Name string
	// DisplayName represents the table header visible in the printed table.
	DisplayName string
}

// GetHeader gets the table header for the current column.
func (c *Column) GetHeader() string {
	header := c.Name

	if c.DisplayName != "" {
		header = c.DisplayName
	}

	header = color.CyanString(header)

	return header
}
