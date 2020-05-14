package table

import (
	"github.com/fatih/color"

	"github.com/giantswarm/gsctl/pkg/sortable"
)

type Column struct {
	sortable.Sortable
	Name        string
	DisplayName string
}

func (c *Column) GetHeader() string {
	header := c.Name

	if c.DisplayName != "" {
		header = c.DisplayName
	}

	header = color.CyanString(header)

	return header
}
