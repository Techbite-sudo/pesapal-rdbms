package index

import (
	"fmt"
	"sync"
)

// Manager manages indexes for tables
type Manager struct {
	indexes map[string]map[string]*BTree // tableName -> columnName -> BTree
	mu      sync.RWMutex
}

// NewManager creates a new index manager
func NewManager() *Manager {
	return &Manager{
		indexes: make(map[string]map[string]*BTree),
	}
}

// CreateIndex creates an index on a table column
func (m *Manager) CreateIndex(tableName, columnName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.indexes[tableName]; !exists {
		m.indexes[tableName] = make(map[string]*BTree)
	}

	if _, exists := m.indexes[tableName][columnName]; exists {
		return fmt.Errorf("index on %s.%s already exists", tableName, columnName)
	}

	m.indexes[tableName][columnName] = NewBTree()
	return nil
}

// DropIndex drops an index
func (m *Manager) DropIndex(tableName, columnName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.indexes[tableName]; !exists {
		return fmt.Errorf("no indexes for table %s", tableName)
	}

	if _, exists := m.indexes[tableName][columnName]; !exists {
		return fmt.Errorf("index on %s.%s does not exist", tableName, columnName)
	}

	delete(m.indexes[tableName], columnName)
	return nil
}

// DropTableIndexes drops all indexes for a table
func (m *Manager) DropTableIndexes(tableName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.indexes, tableName)
}

// Insert inserts a value into an index
func (m *Manager) Insert(tableName, columnName string, key interface{}, rowIndex int) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.indexes[tableName]; !exists {
		return nil // No indexes for this table
	}

	if btree, exists := m.indexes[tableName][columnName]; exists {
		return btree.Insert(key, rowIndex)
	}

	return nil // No index for this column
}

// Search searches for a key in an index
func (m *Manager) Search(tableName, columnName string, key interface{}) (int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.indexes[tableName]; !exists {
		return -1, false
	}

	if btree, exists := m.indexes[tableName][columnName]; exists {
		return btree.Search(key)
	}

	return -1, false
}

// Delete deletes a key from an index
func (m *Manager) Delete(tableName, columnName string, key interface{}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.indexes[tableName]; !exists {
		return false
	}

	if btree, exists := m.indexes[tableName][columnName]; exists {
		return btree.Delete(key)
	}

	return false
}

// HasIndex checks if an index exists
func (m *Manager) HasIndex(tableName, columnName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.indexes[tableName]; !exists {
		return false
	}

	_, exists := m.indexes[tableName][columnName]
	return exists
}

// GetIndexedColumns returns all indexed columns for a table
func (m *Manager) GetIndexedColumns(tableName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.indexes[tableName]; !exists {
		return []string{}
	}

	columns := []string{}
	for col := range m.indexes[tableName] {
		columns = append(columns, col)
	}
	return columns
}
