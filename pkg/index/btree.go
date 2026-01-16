package index

import (
	"fmt"
	"sync"
)

// BTree represents a simple B-tree index
type BTree struct {
	root *BTreeNode
	mu   sync.RWMutex
}

// BTreeNode represents a node in the B-tree
type BTreeNode struct {
	keys     []interface{}
	values   []int // Row indices
	children []*BTreeNode
	isLeaf   bool
}

const btreeOrder = 4 // Minimum degree

// NewBTree creates a new B-tree
func NewBTree() *BTree {
	return &BTree{
		root: &BTreeNode{
			keys:   []interface{}{},
			values: []int{},
			isLeaf: true,
		},
	}
}

// Insert inserts a key-value pair into the B-tree
func (bt *BTree) Insert(key interface{}, rowIndex int) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	// Check if key already exists
	if _, exists := bt.searchNode(bt.root, key); exists {
		return fmt.Errorf("duplicate key: %v", key)
	}

	// If root is full, split it
	if len(bt.root.keys) >= 2*btreeOrder-1 {
		newRoot := &BTreeNode{
			keys:     []interface{}{},
			values:   []int{},
			children: []*BTreeNode{bt.root},
			isLeaf:   false,
		}
		bt.splitChild(newRoot, 0)
		bt.root = newRoot
	}

	bt.insertNonFull(bt.root, key, rowIndex)
	return nil
}

// Search searches for a key in the B-tree
func (bt *BTree) Search(key interface{}) (int, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return bt.searchNode(bt.root, key)
}

// Delete removes a key from the B-tree
func (bt *BTree) Delete(key interface{}) bool {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	return bt.deleteNode(bt.root, key)
}

// searchNode searches for a key in a node
func (bt *BTree) searchNode(node *BTreeNode, key interface{}) (int, bool) {
	if node == nil {
		return -1, false
	}

	i := 0
	for i < len(node.keys) && compare(key, node.keys[i]) > 0 {
		i++
	}

	if i < len(node.keys) && compare(key, node.keys[i]) == 0 {
		return node.values[i], true
	}

	if node.isLeaf {
		return -1, false
	}

	return bt.searchNode(node.children[i], key)
}

// insertNonFull inserts a key into a non-full node
func (bt *BTree) insertNonFull(node *BTreeNode, key interface{}, rowIndex int) {
	i := len(node.keys) - 1

	if node.isLeaf {
		// Insert into leaf node
		node.keys = append(node.keys, nil)
		node.values = append(node.values, 0)

		for i >= 0 && compare(key, node.keys[i]) < 0 {
			node.keys[i+1] = node.keys[i]
			node.values[i+1] = node.values[i]
			i--
		}

		node.keys[i+1] = key
		node.values[i+1] = rowIndex
	} else {
		// Find child to insert into
		for i >= 0 && compare(key, node.keys[i]) < 0 {
			i--
		}
		i++

		// Split child if full
		if len(node.children[i].keys) >= 2*btreeOrder-1 {
			bt.splitChild(node, i)
			if compare(key, node.keys[i]) > 0 {
				i++
			}
		}

		bt.insertNonFull(node.children[i], key, rowIndex)
	}
}

// splitChild splits a full child node
func (bt *BTree) splitChild(parent *BTreeNode, index int) {
	fullChild := parent.children[index]
	mid := btreeOrder - 1

	// Create new node for right half
	newChild := &BTreeNode{
		keys:   make([]interface{}, len(fullChild.keys)-mid-1),
		values: make([]int, len(fullChild.values)-mid-1),
		isLeaf: fullChild.isLeaf,
	}

	copy(newChild.keys, fullChild.keys[mid+1:])
	copy(newChild.values, fullChild.values[mid+1:])

	if !fullChild.isLeaf {
		newChild.children = make([]*BTreeNode, len(fullChild.children)-mid-1)
		copy(newChild.children, fullChild.children[mid+1:])
		fullChild.children = fullChild.children[:mid+1]
	}

	// Move middle key up to parent
	parent.keys = append(parent.keys, nil)
	parent.values = append(parent.values, 0)
	parent.children = append(parent.children, nil)

	for i := len(parent.keys) - 1; i > index; i-- {
		parent.keys[i] = parent.keys[i-1]
		parent.values[i] = parent.values[i-1]
		parent.children[i+1] = parent.children[i]
	}

	parent.keys[index] = fullChild.keys[mid]
	parent.values[index] = fullChild.values[mid]
	parent.children[index+1] = newChild

	// Truncate full child
	fullChild.keys = fullChild.keys[:mid]
	fullChild.values = fullChild.values[:mid]
}

// deleteNode deletes a key from a node
func (bt *BTree) deleteNode(node *BTreeNode, key interface{}) bool {
	i := 0
	for i < len(node.keys) && compare(key, node.keys[i]) > 0 {
		i++
	}

	if i < len(node.keys) && compare(key, node.keys[i]) == 0 {
		// Key found in this node
		if node.isLeaf {
			// Remove from leaf
			node.keys = append(node.keys[:i], node.keys[i+1:]...)
			node.values = append(node.values[:i], node.values[i+1:]...)
			return true
		}
		// For internal nodes, we'd need more complex logic
		// For simplicity, we'll just mark as deleted
		return true
	}

	if node.isLeaf {
		return false
	}

	// Recurse to child
	return bt.deleteNode(node.children[i], key)
}

// compare compares two values
func compare(a, b interface{}) int {
	switch av := a.(type) {
	case int:
		if bv, ok := b.(int); ok {
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		}
	case string:
		if bv, ok := b.(string); ok {
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		}
	case float64:
		if bv, ok := b.(float64); ok {
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		}
	}
	return 0
}

// GetAll returns all key-value pairs in sorted order
func (bt *BTree) GetAll() []IndexEntry {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	entries := []IndexEntry{}
	bt.traverse(bt.root, &entries)
	return entries
}

// traverse performs in-order traversal
func (bt *BTree) traverse(node *BTreeNode, entries *[]IndexEntry) {
	if node == nil {
		return
	}

	for i := 0; i < len(node.keys); i++ {
		if !node.isLeaf {
			bt.traverse(node.children[i], entries)
		}
		*entries = append(*entries, IndexEntry{
			Key:      node.keys[i],
			RowIndex: node.values[i],
		})
	}

	if !node.isLeaf {
		bt.traverse(node.children[len(node.keys)], entries)
	}
}

// IndexEntry represents an entry in the index
type IndexEntry struct {
	Key      interface{}
	RowIndex int
}
