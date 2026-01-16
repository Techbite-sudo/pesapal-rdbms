package storage

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Techbite-sudo/pesapal-rdbms/pkg/index"
)

// Storage manages database storage
type Storage struct {
	dataDir     string
	tables      map[string]*Table
	indexMgr    *index.Manager
	mu          sync.RWMutex
}

// Table represents a database table
type Table struct {
	Schema *Schema
	Rows   []*Row
	mu     sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	s := &Storage{
		dataDir:  dataDir,
		tables:   make(map[string]*Table),
		indexMgr: index.NewManager(),
	}

	// Load existing tables
	if err := s.loadTables(); err != nil {
		return nil, fmt.Errorf("failed to load tables: %w", err)
	}

	return s, nil
}

// CreateTable creates a new table
func (s *Storage) CreateTable(schema *Schema) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tables[schema.TableName]; exists {
		return fmt.Errorf("table %s already exists", schema.TableName)
	}

	table := &Table{
		Schema: schema,
		Rows:   []*Row{},
	}

	s.tables[schema.TableName] = table

	// Create indexes for PRIMARY KEY and UNIQUE columns
	for _, col := range schema.Columns {
		if col.PrimaryKey || col.Unique {
			if err := s.indexMgr.CreateIndex(schema.TableName, col.Name); err != nil {
				delete(s.tables, schema.TableName)
				return fmt.Errorf("failed to create index: %w", err)
			}
		}
	}

	// Persist to disk
	if err := s.saveTable(table); err != nil {
		delete(s.tables, schema.TableName)
		return fmt.Errorf("failed to save table: %w", err)
	}

	return nil
}

// DropTable drops a table
func (s *Storage) DropTable(tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tables[tableName]; !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	delete(s.tables, tableName)

	// Drop all indexes for this table
	s.indexMgr.DropTableIndexes(tableName)

	// Remove from disk
	filePath := s.getTableFilePath(tableName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove table file: %w", err)
	}

	return nil
}

// GetTable returns a table by name
func (s *Storage) GetTable(tableName string) (*Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	table, exists := s.tables[tableName]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	return table, nil
}

// TableExists checks if a table exists
func (s *Storage) TableExists(tableName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.tables[tableName]
	return exists
}

// ListTables returns all table names
func (s *Storage) ListTables() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tables := make([]string, 0, len(s.tables))
	for name := range s.tables {
		tables = append(tables, name)
	}
	return tables
}

// InsertRow inserts a row into a table
func (t *Table) InsertRow(row *Row) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate row length
	if len(row.Values) != len(t.Schema.Columns) {
		return fmt.Errorf("row has %d values but table has %d columns", len(row.Values), len(t.Schema.Columns))
	}

	// Validate each value
	for i, col := range t.Schema.Columns {
		if err := ValidateValue(row.Values[i], col); err != nil {
			return err
		}
	}

	// Check primary key uniqueness
	for _, pkCol := range t.Schema.PrimaryKeys {
		pkIndex := t.Schema.GetColumnIndex(pkCol)
		if pkIndex == -1 {
			continue
		}
		pkValue := row.Values[pkIndex]
		for _, existingRow := range t.Rows {
			if existingRow.Values[pkIndex] == pkValue {
				return fmt.Errorf("duplicate primary key value: %v", pkValue)
			}
		}
	}

	// Check unique constraints
	for _, uniqueCol := range t.Schema.UniqueKeys {
		uniqueIndex := t.Schema.GetColumnIndex(uniqueCol)
		if uniqueIndex == -1 {
			continue
		}
		uniqueValue := row.Values[uniqueIndex]
		if uniqueValue == nil {
			continue // NULL values are allowed in unique columns
		}
		for _, existingRow := range t.Rows {
			if existingRow.Values[uniqueIndex] == uniqueValue {
				return fmt.Errorf("duplicate unique key value in column %s: %v", uniqueCol, uniqueValue)
			}
		}
	}

	t.Rows = append(t.Rows, row)
	return nil
}

// SelectRows returns all rows from a table
func (t *Table) SelectRows() []*Row {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	rows := make([]*Row, len(t.Rows))
	copy(rows, t.Rows)
	return rows
}

// UpdateRows updates rows matching a condition
func (t *Table) UpdateRows(condition func(*Row) bool, updates map[string]interface{}) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := 0
	for _, row := range t.Rows {
		if condition == nil || condition(row) {
			for colName, value := range updates {
				colIndex := t.Schema.GetColumnIndex(colName)
				if colIndex == -1 {
					return count, fmt.Errorf("column %s not found", colName)
				}

				col := t.Schema.Columns[colIndex]
				if err := ValidateValue(value, col); err != nil {
					return count, err
				}

				row.Values[colIndex] = value
			}
			count++
		}
	}

	return count, nil
}

// DeleteRows deletes rows matching a condition
func (t *Table) DeleteRows(condition func(*Row) bool) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if condition == nil {
		// Delete all rows
		count := len(t.Rows)
		t.Rows = []*Row{}
		return count
	}

	newRows := []*Row{}
	count := 0
	for _, row := range t.Rows {
		if !condition(row) {
			newRows = append(newRows, row)
		} else {
			count++
		}
	}

	t.Rows = newRows
	return count
}

// getTableFilePath returns the file path for a table
func (s *Storage) getTableFilePath(tableName string) string {
	return filepath.Join(s.dataDir, tableName+".tbl")
}

// saveTable saves a table to disk
func (s *Storage) saveTable(table *Table) error {
	filePath := s.getTableFilePath(table.Schema.TableName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(table.Schema); err != nil {
		return err
	}
	if err := encoder.Encode(table.Rows); err != nil {
		return err
	}

	return nil
}

// SaveAllTables saves all tables to disk
func (s *Storage) SaveAllTables() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, table := range s.tables {
		if err := s.saveTable(table); err != nil {
			return fmt.Errorf("failed to save table %s: %w", table.Schema.TableName, err)
		}
	}

	return nil
}

// loadTables loads all tables from disk
func (s *Storage) loadTables() error {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No tables yet
		}
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".tbl" {
			continue
		}

		tableName := file.Name()[:len(file.Name())-4]
		table, err := s.loadTable(tableName)
		if err != nil {
			return fmt.Errorf("failed to load table %s: %w", tableName, err)
		}

		s.tables[tableName] = table
	}

	return nil
}

// loadTable loads a single table from disk
func (s *Storage) loadTable(tableName string) (*Table, error) {
	filePath := s.getTableFilePath(tableName)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)

	var schema Schema
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	var rows []*Row
	if err := decoder.Decode(&rows); err != nil {
		return nil, err
	}

	return &Table{
		Schema: &schema,
		Rows:   rows,
	}, nil
}
