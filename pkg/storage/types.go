package storage

import (
	"fmt"
)

// DataType represents column data types
type DataType int

const (
	TypeInteger DataType = iota
	TypeVarchar
	TypeBoolean
	TypeFloat
)

// String returns string representation of data type
func (d DataType) String() string {
	switch d {
	case TypeInteger:
		return "INTEGER"
	case TypeVarchar:
		return "VARCHAR"
	case TypeBoolean:
		return "BOOLEAN"
	case TypeFloat:
		return "FLOAT"
	default:
		return "UNKNOWN"
	}
}

// Column represents a table column definition
type Column struct {
	Name       string
	DataType   DataType
	Size       int  // for VARCHAR
	PrimaryKey bool
	Unique     bool
	NotNull    bool
}

// Schema represents a table schema
type Schema struct {
	TableName   string
	Columns     []Column
	PrimaryKeys []string
	UniqueKeys  []string
}

// NewSchema creates a new schema
func NewSchema(tableName string) *Schema {
	return &Schema{
		TableName:   tableName,
		Columns:     []Column{},
		PrimaryKeys: []string{},
		UniqueKeys:  []string{},
	}
}

// AddColumn adds a column to the schema
func (s *Schema) AddColumn(col Column) {
	s.Columns = append(s.Columns, col)
	if col.PrimaryKey {
		s.PrimaryKeys = append(s.PrimaryKeys, col.Name)
	}
	if col.Unique {
		s.UniqueKeys = append(s.UniqueKeys, col.Name)
	}
}

// GetColumn returns a column by name
func (s *Schema) GetColumn(name string) (*Column, error) {
	for i := range s.Columns {
		if s.Columns[i].Name == name {
			return &s.Columns[i], nil
		}
	}
	return nil, fmt.Errorf("column %s not found", name)
}

// GetColumnIndex returns the index of a column by name
func (s *Schema) GetColumnIndex(name string) int {
	for i, col := range s.Columns {
		if col.Name == name {
			return i
		}
	}
	return -1
}

// Row represents a single row of data
type Row struct {
	Values []interface{}
}

// NewRow creates a new row
func NewRow(values []interface{}) *Row {
	return &Row{Values: values}
}

// Get returns the value at the given index
func (r *Row) Get(index int) interface{} {
	if index < 0 || index >= len(r.Values) {
		return nil
	}
	return r.Values[index]
}

// Set sets the value at the given index
func (r *Row) Set(index int, value interface{}) {
	if index >= 0 && index < len(r.Values) {
		r.Values[index] = value
	}
}

// ValidateValue validates a value against a column definition
func ValidateValue(value interface{}, col Column) error {
	if value == nil {
		if col.NotNull {
			return fmt.Errorf("column %s cannot be NULL", col.Name)
		}
		return nil
	}

	switch col.DataType {
	case TypeInteger:
		if _, ok := value.(int); !ok {
			return fmt.Errorf("column %s expects INTEGER, got %T", col.Name, value)
		}
	case TypeVarchar:
		if str, ok := value.(string); ok {
			if col.Size > 0 && len(str) > col.Size {
				return fmt.Errorf("column %s: string length %d exceeds maximum %d", col.Name, len(str), col.Size)
			}
		} else {
			return fmt.Errorf("column %s expects VARCHAR, got %T", col.Name, value)
		}
	case TypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("column %s expects BOOLEAN, got %T", col.Name, value)
		}
	case TypeFloat:
		switch value.(type) {
		case float32, float64:
			// OK
		default:
			return fmt.Errorf("column %s expects FLOAT, got %T", col.Name, value)
		}
	}

	return nil
}
