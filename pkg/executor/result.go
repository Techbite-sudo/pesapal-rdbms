package executor

import (
	"fmt"
	"strings"
)

// Result represents the result of a SQL query execution
type Result struct {
	Columns      []string        // Column names for SELECT queries
	Rows         [][]interface{} // Row data for SELECT queries
	Message      string          // Message for non-SELECT queries
	RowsAffected int             // Number of rows affected
}

// FormatTable formats the result as a table string
func (r *Result) FormatTable() string {
	if len(r.Columns) == 0 {
		return r.Message
	}

	if len(r.Rows) == 0 {
		return "No rows returned."
	}

	// Calculate column widths
	widths := make([]int, len(r.Columns))
	for i, col := range r.Columns {
		widths[i] = len(col)
	}

	for _, row := range r.Rows {
		for i, val := range row {
			valStr := formatValue(val)
			if len(valStr) > widths[i] {
				widths[i] = len(valStr)
			}
		}
	}

	var sb strings.Builder

	// Top border
	sb.WriteString("+")
	for _, width := range widths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("+")
	}
	sb.WriteString("\n")

	// Header
	sb.WriteString("|")
	for i, col := range r.Columns {
		sb.WriteString(" ")
		sb.WriteString(padRight(col, widths[i]))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Header separator
	sb.WriteString("+")
	for _, width := range widths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("+")
	}
	sb.WriteString("\n")

	// Rows
	for _, row := range r.Rows {
		sb.WriteString("|")
		for i, val := range row {
			sb.WriteString(" ")
			sb.WriteString(padRight(formatValue(val), widths[i]))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString("+")
	for _, width := range widths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("+")
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("\n%d row(s) returned.\n", len(r.Rows)))

	return sb.String()
}

// formatValue formats a value for display
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", val)
}

// padRight pads a string to the right with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
