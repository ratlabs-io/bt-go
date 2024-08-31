// Package bt provides a synchronized context for sharing data between actions in a behavior tree.
package bt

import (
	"context"
	"sync"
)

// BehaviorContext defines the interface for a synchronized context.
type BehaviorContext interface {
	Set(key, value interface{})
	Get(key interface{}) (value interface{}, ok bool)
	Delete(key interface{})
}

// behaviorContextImpl is the concrete implementation of the BehaviorContext interface.
type behaviorContextImpl struct {
	Ctx         context.Context
	mu          sync.RWMutex
	ContextData map[interface{}]interface{}
}

// NewBehaviorContext creates a new BehaviorContext with the specified base context.
func NewBehaviorContext(ctx context.Context) BehaviorContext {
	return &behaviorContextImpl{
		Ctx:         ctx,
		ContextData: make(map[interface{}]interface{}),
	}
}

// Set stores the value in the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Set(key, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.ContextData[key] = value
}

// Get retrieves the value from the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Get(key interface{}) (value interface{}, ok bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	value, ok = bc.ContextData[key]
	return
}

// Delete removes the value from the behaviorContextImpl for the given key.
func (bc *behaviorContextImpl) Delete(key interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.ContextData, key)
}
