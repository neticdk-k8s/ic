package ui

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// NewTable creates a new table with default settings
func NewTable(writer io.Writer, headers []string) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	if len(headers) > 0 {
		table.SetHeader(headers)
	} else {
		table.SetAutoFormatHeaders(false)
	}
	table.SetAutoWrapText(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	return table
}
