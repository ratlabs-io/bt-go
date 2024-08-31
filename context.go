// Package bt provides a synchronized context for sharing data between actions in a behavior tree.
package bt

import (
	"context"
	"sync"
)

// BehaviorContext defines the interface for a synchronized context.
type BehaviorContext interface {
	Set(key string, value interface{})
	Get(key string) (value interface{}, ok bool)
	Delete(key string)
}

// behaviorContextImpl is the concrete implementation of the BehaviorContext interface.
type behaviorContextImpl struct {
	Ctx         context.Context
	mu          sync.RWMutex
	ContextData map[string]interface{}
}

// NewBehaviorContext creates a new BehaviorContext with the specified base context.
func NewBehaviorContext(ctx context.Context) BehaviorContext {
	return &behaviorContextImpl{
		Ctx:         ctx,
		ContextData: make(map[string]interface{}),
	}
}

// Set stores the value in the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Set(key string, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.ContextData[key] = value
}

// Get retrieves the value from the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Get(key string) (value interface{}, ok bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	value, ok = bc.ContextData[key]
	return
}

// Delete removes the value from the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Delete(key string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.ContextData, key)
}
