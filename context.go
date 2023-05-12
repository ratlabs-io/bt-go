// Package bt provides a synchronized context for sharing data between actions in a behavior tree.
package bt

import (
	"context"
	"sync"
)

// BehaviorContext provides a synchronized context for sharing data between actions in a behavior tree.
type BehaviorContext struct {
	Ctx         context.Context
	mu          sync.RWMutex
	ContextData map[interface{}]interface{}
}

// NewBehaviorContext creates a new BehaviorContext with the specified base context.
func NewBehaviorContext(ctx context.Context) *BehaviorContext {
	return &BehaviorContext{
		Ctx:         ctx,
		ContextData: make(map[interface{}]interface{}),
	}
}

// Set stores the value in the BehaviorContext for the given key.
func (bc *BehaviorContext) Set(key, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.ContextData[key] = value
}

// Get retrieves the value from the BehaviorContext for the given key.
// It returns the value and a boolean indicating whether the key exists in the BehaviorContext.
func (bc *BehaviorContext) Get(key interface{}) (value interface{}, ok bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	value, ok = bc.ContextData[key]
	return
}

// Delete removes the value from the BehaviorContext for the given key.
func (bc *BehaviorContext) Delete(key interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.ContextData, key)
}
