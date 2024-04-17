package ui

import (
	"fmt"
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

// RenderKVTable creates a new key/value table with default settings
func RenderKVTable(writer io.Writer, title string, rows [][]string) {
	table := NewTable(writer, []string{})
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(false)
	table.AppendBulk(rows)
	fmt.Fprintf(writer, "%s:\n", title)
	table.Render()
}
