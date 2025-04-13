package bt

import (
	"sync"
)

// Blackboard provides a hierarchical, thread-safe key-value store for behavior trees.
// It can have parent blackboards for forming scopes of data.
type Blackboard struct {
	mu      sync.RWMutex
	data    map[string]interface{}
	parent  *Blackboard
	entries map[string]bool // Records keys that are specifically set on this blackboard
}

// NewBlackboard creates a new Blackboard instance.
func NewBlackboard() *Blackboard {
	return &Blackboard{
		data:    make(map[string]interface{}),
		entries: make(map[string]bool),
	}
}

// NewBlackboardWithParent creates a new Blackboard with a parent blackboard.
func NewBlackboardWithParent(parent *Blackboard) *Blackboard {
	bb := NewBlackboard()
	bb.parent = parent
	return bb
}

// Get retrieves a value from the blackboard.
// It first checks the current blackboard, then falls back to parent blackboards.
func (bb *Blackboard) Get(key string) (interface{}, bool) {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	// Check if the key exists in the current blackboard
	if value, ok := bb.data[key]; ok {
		return value, true
	}

	// If not found and we have a parent, check the parent
	if bb.parent != nil {
		return bb.parent.Get(key)
	}

	// Not found
	return nil, false
}

// Set stores a value in the blackboard.
func (bb *Blackboard) Set(key string, value interface{}) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	bb.data[key] = value
	bb.entries[key] = true
}

// HasLocal checks if a key exists in this blackboard (not checking parents).
func (bb *Blackboard) HasLocal(key string) bool {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	_, exists := bb.entries[key]
	return exists
}

// Has checks if a key exists in this blackboard or any parent.
func (bb *Blackboard) Has(key string) bool {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	if _, ok := bb.data[key]; ok {
		return true
	}

	if bb.parent != nil {
		return bb.parent.Has(key)
	}

	return false
}

// Delete removes a key from this blackboard.
// It does not affect parent blackboards.
func (bb *Blackboard) Delete(key string) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	delete(bb.data, key)
	delete(bb.entries, key)
}

// Clear removes all entries from this blackboard.
// It does not affect parent blackboards.
func (bb *Blackboard) Clear() {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	bb.data = make(map[string]interface{})
	bb.entries = make(map[string]bool)
}

// Entries returns all keys that are specifically set on this blackboard.
func (bb *Blackboard) Entries() []string {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	keys := make([]string, 0, len(bb.entries))
	for k := range bb.entries {
		keys = append(keys, k)
	}
	return keys
}

// AllEntries returns all keys accessible from this blackboard,
// including those from parent blackboards.
func (bb *Blackboard) AllEntries() []string {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	// Start with our own entries
	allKeys := make(map[string]bool)
	for k := range bb.entries {
		allKeys[k] = true
	}

	// Add parent entries
	if bb.parent != nil {
		parentKeys := bb.parent.AllEntries()
		for _, k := range parentKeys {
			allKeys[k] = true
		}
	}

	// Convert to slice
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	return keys
}

// WithBlackboard is an option for BehaviorContext to set a blackboard.
func WithBlackboard(bb *Blackboard) func(ctx *behaviorContextImpl) {
	return func(ctx *behaviorContextImpl) {
		ctx.Blackboard = bb
	}
}
