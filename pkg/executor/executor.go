package executor

import (
	"fmt"
	"strings"

	"github.com/Techbite-sudo/pesapal-rdbms/pkg/parser"
	"github.com/Techbite-sudo/pesapal-rdbms/pkg/storage"
)

// Executor executes SQL statements
type Executor struct {
	storage *storage.Storage
}

// NewExecutor creates a new executor
func NewExecutor(storage *storage.Storage) *Executor {
	return &Executor{storage: storage}
}

// Execute executes a SQL statement
func (e *Executor) Execute(stmt parser.Statement) (*Result, error) {
	switch s := stmt.(type) {
	case *parser.CreateTableStmt:
		return e.executeCreateTable(s)
	case *parser.DropTableStmt:
		return e.executeDropTable(s)
	case *parser.InsertStmt:
		return e.executeInsert(s)
	case *parser.SelectStmt:
		return e.executeSelect(s)
	case *parser.UpdateStmt:
		return e.executeUpdate(s)
	case *parser.DeleteStmt:
		return e.executeDelete(s)
	default:
		return nil, fmt.Errorf("unsupported statement type")
	}
}

// executeCreateTable executes CREATE TABLE statement
func (e *Executor) executeCreateTable(stmt *parser.CreateTableStmt) (*Result, error) {
	schema := storage.NewSchema(stmt.TableName)

	for _, colDef := range stmt.Columns {
		col := storage.Column{
			Name:       colDef.Name,
			Size:       colDef.Size,
			PrimaryKey: colDef.PrimaryKey,
			Unique:     colDef.Unique,
			NotNull:    colDef.NotNull,
		}

		// Convert data type
		switch strings.ToUpper(colDef.DataType) {
		case "INTEGER":
			col.DataType = storage.TypeInteger
		case "VARCHAR":
			col.DataType = storage.TypeVarchar
		case "BOOLEAN":
			col.DataType = storage.TypeBoolean
		case "FLOAT":
			col.DataType = storage.TypeFloat
		default:
			return nil, fmt.Errorf("unsupported data type: %s", colDef.DataType)
		}

		schema.AddColumn(col)
	}

	if err := e.storage.CreateTable(schema); err != nil {
		return nil, err
	}

	// Save to disk
	if err := e.storage.SaveAllTables(); err != nil {
		return nil, fmt.Errorf("failed to persist table: %w", err)
	}

	return &Result{
		Message:      fmt.Sprintf("Table '%s' created successfully", stmt.TableName),
		RowsAffected: 0,
	}, nil
}

// executeDropTable executes DROP TABLE statement
func (e *Executor) executeDropTable(stmt *parser.DropTableStmt) (*Result, error) {
	if err := e.storage.DropTable(stmt.TableName); err != nil {
		return nil, err
	}

	return &Result{
		Message:      fmt.Sprintf("Table '%s' dropped successfully", stmt.TableName),
		RowsAffected: 0,
	}, nil
}

// executeInsert executes INSERT statement
func (e *Executor) executeInsert(stmt *parser.InsertStmt) (*Result, error) {
	table, err := e.storage.GetTable(stmt.TableName)
	if err != nil {
		return nil, err
	}

	// Determine column order
	columns := stmt.Columns
	if len(columns) == 0 {
		// Use all columns in schema order
		for _, col := range table.Schema.Columns {
			columns = append(columns, col.Name)
		}
	}

	// Validate columns exist
	columnIndices := make([]int, len(columns))
	for i, colName := range columns {
		idx := table.Schema.GetColumnIndex(colName)
		if idx == -1 {
			return nil, fmt.Errorf("column %s does not exist in table %s", colName, stmt.TableName)
		}
		columnIndices[i] = idx
	}

	rowsInserted := 0
	for _, valueSet := range stmt.Values {
		if len(valueSet) != len(columns) {
			return nil, fmt.Errorf("column count mismatch: expected %d, got %d", len(columns), len(valueSet))
		}

		// Create a row with NULL values
		row := storage.NewRow(make([]interface{}, len(table.Schema.Columns)))
		for i := range row.Values {
			row.Values[i] = nil
		}

		// Fill in provided values
		for i, expr := range valueSet {
			value, err := e.evaluateExpression(expr, nil)
			if err != nil {
				return nil, err
			}
			row.Values[columnIndices[i]] = value
		}

		if err := table.InsertRow(row); err != nil {
			return nil, err
		}
		rowsInserted++
	}

	// Save to disk
	if err := e.storage.SaveAllTables(); err != nil {
		return nil, fmt.Errorf("failed to persist data: %w", err)
	}

	return &Result{
		Message:      fmt.Sprintf("%d row(s) inserted", rowsInserted),
		RowsAffected: rowsInserted,
	}, nil
}

// executeSelect executes SELECT statement
func (e *Executor) executeSelect(stmt *parser.SelectStmt) (*Result, error) {
	table, err := e.storage.GetTable(stmt.TableName)
	if err != nil {
		return nil, err
	}

	// Get all rows from the main table
	rows := table.SelectRows()

	// Handle JOINs
	if len(stmt.Joins) > 0 {
		return e.executeSelectWithJoin(stmt, table, rows)
	}

	// Filter by WHERE clause (no joins)
	if stmt.Where != nil {
		filteredRows := []*storage.Row{}
		for _, row := range rows {
			match, err := e.evaluateCondition(stmt.Where, row, table.Schema)
			if err != nil {
				return nil, err
			}
			if match {
				filteredRows = append(filteredRows, row)
			}
		}
		rows = filteredRows
	}

	// Determine columns to return
	var columnIndices []int
	var columnNames []string

	if len(stmt.Columns) == 1 && stmt.Columns[0] == "*" {
		// Select all columns
		for i, col := range table.Schema.Columns {
			columnIndices = append(columnIndices, i)
			columnNames = append(columnNames, col.Name)
		}
	} else {
		// Select specific columns
		for _, colName := range stmt.Columns {
			idx := table.Schema.GetColumnIndex(colName)
			if idx == -1 {
				return nil, fmt.Errorf("column %s does not exist", colName)
			}
			columnIndices = append(columnIndices, idx)
			columnNames = append(columnNames, colName)
		}
	}

	// Build result rows
	resultRows := [][]interface{}{}
	for _, row := range rows {
		resultRow := []interface{}{}
		for _, idx := range columnIndices {
			resultRow = append(resultRow, row.Values[idx])
		}
		resultRows = append(resultRows, resultRow)
	}

	return &Result{
		Columns:      columnNames,
		Rows:         resultRows,			RowsAffected: len(resultRows),
	}, nil
}

// executeSelectWithJoin executes SELECT with JOIN
func (e *Executor) executeSelectWithJoin(stmt *parser.SelectStmt, leftTable *storage.Table, leftRows []*storage.Row) (*Result, error) {
	// For now, we only support INNER JOIN with one join table
	if len(stmt.Joins) > 1 {
		return nil, fmt.Errorf("multiple joins not yet supported")
	}

	join := stmt.Joins[0]
	rightTable, err := e.storage.GetTable(join.TableName)
	if err != nil {
		return nil, err
	}

	rightRows := rightTable.SelectRows()

	// Perform nested loop join
	joinedRows := [][]interface{}{}

	for _, leftRow := range leftRows {
		for _, rightRow := range rightRows {
			// Create a combined row
			combinedRow := &CombinedRow{
				leftRow:    leftRow,
				rightRow:   rightRow,
				leftSchema: leftTable.Schema,
				rightSchema: rightTable.Schema,
				leftTableName: stmt.TableName,
				rightTableName: join.TableName,
			}

			// Evaluate join condition
			if join.On != nil {
				match, err := e.evaluateJoinCondition(join.On, combinedRow)
				if err != nil {
					return nil, err
				}
				if !match {
					continue
				}
			}

			// Apply WHERE clause if present
			if stmt.Where != nil {
				match, err := e.evaluateJoinCondition(stmt.Where, combinedRow)
				if err != nil {
					return nil, err
				}
				if !match {
					continue
				}
			}

			// Combine rows
			combined := append([]interface{}{}, leftRow.Values...)
			combined = append(combined, rightRow.Values...)
			joinedRows = append(joinedRows, combined)
		}
	}

	// Determine columns to return
	var columnIndices []int
	var columnNames []string

	if len(stmt.Columns) == 1 && stmt.Columns[0] == "*" {
		// Select all columns from both tables
		for i, col := range leftTable.Schema.Columns {
			columnIndices = append(columnIndices, i)
			columnNames = append(columnNames, stmt.TableName+"."+col.Name)
		}
		for i, col := range rightTable.Schema.Columns {
			columnIndices = append(columnIndices, len(leftTable.Schema.Columns)+i)
			columnNames = append(columnNames, join.TableName+"."+col.Name)
		}
	} else {
		// Select specific columns (support table.column notation)
		for _, colSpec := range stmt.Columns {
			parts := strings.Split(colSpec, ".")
			if len(parts) == 2 {
				// table.column format
				tableName := parts[0]
				colName := parts[1]
				if tableName == stmt.TableName {
					idx := leftTable.Schema.GetColumnIndex(colName)
					if idx == -1 {
						return nil, fmt.Errorf("column %s not found in table %s", colName, tableName)
					}
					columnIndices = append(columnIndices, idx)
					columnNames = append(columnNames, colSpec)
				} else if tableName == join.TableName {
					idx := rightTable.Schema.GetColumnIndex(colName)
					if idx == -1 {
						return nil, fmt.Errorf("column %s not found in table %s", colName, tableName)
					}
					columnIndices = append(columnIndices, len(leftTable.Schema.Columns)+idx)
					columnNames = append(columnNames, colSpec)
				} else {
					return nil, fmt.Errorf("unknown table: %s", tableName)
				}
			} else {
				// Try to find in left table first, then right
				idx := leftTable.Schema.GetColumnIndex(colSpec)
				if idx != -1 {
					columnIndices = append(columnIndices, idx)
					columnNames = append(columnNames, colSpec)
				} else {
					idx = rightTable.Schema.GetColumnIndex(colSpec)
					if idx != -1 {
						columnIndices = append(columnIndices, len(leftTable.Schema.Columns)+idx)
						columnNames = append(columnNames, colSpec)
					} else {
						return nil, fmt.Errorf("column %s not found", colSpec)
					}
				}
			}
		}
	}

	// Build result rows
	resultRows := [][]interface{}{}
	for _, row := range joinedRows {
		resultRow := []interface{}{}
		for _, idx := range columnIndices {
			resultRow = append(resultRow, row[idx])
		}
		resultRows = append(resultRows, resultRow)
	}

	return &Result{
		Columns:      columnNames,
		Rows:         resultRows,
		RowsAffected: len(resultRows),
	}, nil
}

// CombinedRow represents a row from a JOIN operation
type CombinedRow struct {
	leftRow        *storage.Row
	rightRow       *storage.Row
	leftSchema     *storage.Schema
	rightSchema    *storage.Schema
	leftTableName  string
	rightTableName string
}

// evaluateJoinCondition evaluates a condition for a joined row
func (e *Executor) evaluateJoinCondition(expr parser.Expression, row *CombinedRow) (bool, error) {
	switch ex := expr.(type) {
	case *parser.BinaryExpr:
		left, err := e.getJoinColumnValue(ex.Left, row)
		if err != nil {
			return false, err
		}

		right, err := e.getJoinColumnValue(ex.Right, row)
		if err != nil {
			return false, err
		}

		return e.compareValues(left, right, ex.Operator)
	default:
		return false, fmt.Errorf("unsupported join condition type")
	}
}

// getJoinColumnValue gets a value from a joined row
func (e *Executor) getJoinColumnValue(expr parser.Expression, row *CombinedRow) (interface{}, error) {
	switch ex := expr.(type) {
	case *parser.Identifier:
		// Check if it's table.column format
		parts := strings.Split(ex.Value, ".")
		if len(parts) == 2 {
			tableName := parts[0]
			colName := parts[1]
			if tableName == row.leftTableName {
				idx := row.leftSchema.GetColumnIndex(colName)
				if idx == -1 {
					return nil, fmt.Errorf("column %s not found in table %s", colName, tableName)
				}
				return row.leftRow.Values[idx], nil
			} else if tableName == row.rightTableName {
				idx := row.rightSchema.GetColumnIndex(colName)
				if idx == -1 {
					return nil, fmt.Errorf("column %s not found in table %s", colName, tableName)
				}
				return row.rightRow.Values[idx], nil
			} else {
				return nil, fmt.Errorf("unknown table: %s", tableName)
			}
		} else {
			// Try to find in left table first
			idx := row.leftSchema.GetColumnIndex(ex.Value)
			if idx != -1 {
				return row.leftRow.Values[idx], nil
			}
			// Try right table
			idx = row.rightSchema.GetColumnIndex(ex.Value)
			if idx != -1 {
				return row.rightRow.Values[idx], nil
			}
			return nil, fmt.Errorf("column %s not found", ex.Value)
		}
	case *parser.Literal:
		return ex.Value, nil
	case *parser.NullLiteral:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported expression in join condition")
	}
}

// executeUpdate executes UPDATE statement
func (e *Executor) executeUpdate(stmt *parser.UpdateStmt) (*Result, error) {
	table, err := e.storage.GetTable(stmt.TableName)
	if err != nil {
		return nil, err
	}

	// Build condition function
	var condition func(*storage.Row) bool
	if stmt.Where != nil {
		condition = func(row *storage.Row) bool {
			match, err := e.evaluateCondition(stmt.Where, row, table.Schema)
			if err != nil {
				return false
			}
			return match
		}
	}

	// Evaluate update values
	updates := make(map[string]interface{})
	for colName, expr := range stmt.Set {
		value, err := e.evaluateExpression(expr, nil)
		if err != nil {
			return nil, err
		}
		updates[colName] = value
	}

	count, err := table.UpdateRows(condition, updates)
	if err != nil {
		return nil, err
	}

	// Save to disk
	if err := e.storage.SaveAllTables(); err != nil {
		return nil, fmt.Errorf("failed to persist data: %w", err)
	}

	return &Result{
		Message:      fmt.Sprintf("%d row(s) updated", count),
		RowsAffected: count,
	}, nil
}

// executeDelete executes DELETE statement
func (e *Executor) executeDelete(stmt *parser.DeleteStmt) (*Result, error) {
	table, err := e.storage.GetTable(stmt.TableName)
	if err != nil {
		return nil, err
	}

	// Build condition function
	var condition func(*storage.Row) bool
	if stmt.Where != nil {
		condition = func(row *storage.Row) bool {
			match, err := e.evaluateCondition(stmt.Where, row, table.Schema)
			if err != nil {
				return false
			}
			return match
		}
	}

	count := table.DeleteRows(condition)

	// Save to disk
	if err := e.storage.SaveAllTables(); err != nil {
		return nil, fmt.Errorf("failed to persist data: %w", err)
	}

	return &Result{
		Message:      fmt.Sprintf("%d row(s) deleted", count),
		RowsAffected: count,
	}, nil
}

// evaluateExpression evaluates an expression to a value
func (e *Executor) evaluateExpression(expr parser.Expression, row *storage.Row) (interface{}, error) {
	switch e := expr.(type) {
	case *parser.Literal:
		return e.Value, nil
	case *parser.NullLiteral:
		return nil, nil
	case *parser.Identifier:
		if row == nil {
			return nil, fmt.Errorf("cannot evaluate identifier without row context")
		}
		return nil, fmt.Errorf("identifier evaluation in INSERT not supported")
	case *parser.BinaryExpr:
		return nil, fmt.Errorf("binary expressions in INSERT not supported")
	default:
		return nil, fmt.Errorf("unsupported expression type")
	}
}

// evaluateCondition evaluates a WHERE condition
func (e *Executor) evaluateCondition(expr parser.Expression, row *storage.Row, schema *storage.Schema) (bool, error) {
	switch ex := expr.(type) {
	case *parser.BinaryExpr:
		left, err := e.getColumnValue(ex.Left, row, schema)
		if err != nil {
			return false, err
		}

		right, err := e.getColumnValue(ex.Right, row, schema)
		if err != nil {
			return false, err
		}

		return e.compareValues(left, right, ex.Operator)
	default:
		return false, fmt.Errorf("unsupported condition type")
	}
}

// getColumnValue gets a value from a row or literal
func (e *Executor) getColumnValue(expr parser.Expression, row *storage.Row, schema *storage.Schema) (interface{}, error) {
	switch ex := expr.(type) {
	case *parser.Identifier:
		idx := schema.GetColumnIndex(ex.Value)
		if idx == -1 {
			return nil, fmt.Errorf("column %s not found", ex.Value)
		}
		return row.Values[idx], nil
	case *parser.Literal:
		return ex.Value, nil
	case *parser.NullLiteral:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported expression in condition")
	}
}

// compareValues compares two values using an operator
func (e *Executor) compareValues(left, right interface{}, operator string) (bool, error) {
	// Handle NULL comparisons
	if left == nil || right == nil {
		if operator == "=" {
			return left == right, nil
		}
		return false, nil
	}

	switch operator {
	case "=":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		return e.lessThan(left, right)
	case ">":
		return e.greaterThan(left, right)
	case "<=":
		return e.lessThanOrEqual(left, right)
	case ">=":
		return e.greaterThanOrEqual(left, right)
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// Comparison helper functions
func (e *Executor) lessThan(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l < r, nil
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l < r, nil
		}
	case string:
		if r, ok := right.(string); ok {
			return l < r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (e *Executor) greaterThan(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l > r, nil
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l > r, nil
		}
	case string:
		if r, ok := right.(string); ok {
			return l > r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (e *Executor) lessThanOrEqual(left, right interface{}) (bool, error) {
	lt, err1 := e.lessThan(left, right)
	if err1 == nil && lt {
		return true, nil
	}
	return left == right, nil
}

func (e *Executor) greaterThanOrEqual(left, right interface{}) (bool, error) {
	gt, err1 := e.greaterThan(left, right)
	if err1 == nil && gt {
		return true, nil
	}
	return left == right, nil
}
