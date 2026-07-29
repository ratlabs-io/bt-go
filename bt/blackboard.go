package bt

import (
	"sync"
)

// Blackboard is a hierarchical, thread-safe key-value store for tree state.
//
// Child blackboards can see parent entries (Get/Has walk upward). Writes and
// deletes apply only to the local blackboard, so a child can shadow a parent
// key without mutating it.
type Blackboard struct {
	mu     sync.RWMutex
	data   map[string]interface{}
	parent *Blackboard
}

// NewBlackboard creates an empty root blackboard.
func NewBlackboard() *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
	}
}

// NewBlackboardWithParent creates a blackboard that falls back to parent on miss.
func NewBlackboardWithParent(parent *Blackboard) *Blackboard {
	return &Blackboard{
		data:   make(map[string]interface{}),
		parent: parent,
	}
}

// Get retrieves a value, checking this blackboard then parents.
func (bb *Blackboard) Get(key string) (interface{}, bool) {
	bb.mu.RLock()
	value, ok := bb.data[key]
	parent := bb.parent
	bb.mu.RUnlock()

	if ok {
		return value, true
	}
	if parent != nil {
		return parent.Get(key)
	}
	return nil, false
}

// Set stores a value on this blackboard only.
func (bb *Blackboard) Set(key string, value interface{}) {
	bb.mu.Lock()
	defer bb.mu.Unlock()
	bb.data[key] = value
}

// HasLocal reports whether key is set on this blackboard (ignoring parents).
func (bb *Blackboard) HasLocal(key string) bool {
	bb.mu.RLock()
	defer bb.mu.RUnlock()
	_, exists := bb.data[key]
	return exists
}

// Has reports whether key exists on this blackboard or any parent.
func (bb *Blackboard) Has(key string) bool {
	bb.mu.RLock()
	_, ok := bb.data[key]
	parent := bb.parent
	bb.mu.RUnlock()

	if ok {
		return true
	}
	if parent != nil {
		return parent.Has(key)
	}
	return false
}

// Delete removes a key from this blackboard only.
func (bb *Blackboard) Delete(key string) {
	bb.mu.Lock()
	defer bb.mu.Unlock()
	delete(bb.data, key)
}

// Clear removes all local entries. Parents are unaffected.
func (bb *Blackboard) Clear() {
	bb.mu.Lock()
	defer bb.mu.Unlock()
	bb.data = make(map[string]interface{})
}

// Entries returns keys set on this blackboard (not parents).
func (bb *Blackboard) Entries() []string {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	keys := make([]string, 0, len(bb.data))
	for k := range bb.data {
		keys = append(keys, k)
	}
	return keys
}

// AllEntries returns keys visible from this blackboard, including parents.
// Local keys shadow parent keys of the same name (counted once).
func (bb *Blackboard) AllEntries() []string {
	bb.mu.RLock()
	allKeys := make(map[string]bool, len(bb.data))
	for k := range bb.data {
		allKeys[k] = true
	}
	parent := bb.parent
	bb.mu.RUnlock()

	if parent != nil {
		for _, k := range parent.AllEntries() {
			allKeys[k] = true
		}
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	return keys
}
