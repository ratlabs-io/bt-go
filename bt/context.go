// Package bt provides a synchronized context for sharing data between actions in a behavior tree.
package bt

import (
	"context"
	"sync"
)

// BehaviorContext defines the interface for a thread-safe context used to share data between behavior tree nodes.
// It allows setting, getting, and deleting key-value pairs in a concurrent-safe manner.
type BehaviorContext interface {
	// Set stores a value for a given key in the context.
	Set(key string, value interface{})
	// Get retrieves a value for a given key from the context, returning the value and a boolean indicating if the key was found.
	Get(key string) (value interface{}, ok bool)
	// Delete removes a key-value pair from the context.
	Delete(key string)
	// Has checks if a key exists in the context without retrieving the value.
	Has(key string) bool
	// GetBlackboard returns the blackboard associated with this context.
	GetBlackboard() *Blackboard
}

// behaviorContextImpl is the concrete implementation of the BehaviorContext interface.
// It wraps a standard context.Context and provides thread-safe access to a map for storing data.
type behaviorContextImpl struct {
	Ctx         context.Context        // Ctx holds the underlying context for cancellation and deadlines.
	mu          sync.RWMutex           // mu ensures thread-safe access to the data map.
	ContextData map[string]interface{} // ContextData stores the key-value pairs shared across the behavior tree.
	Blackboard  *Blackboard            // Blackboard provides hierarchical data sharing capabilities.
}

// ContextOption is a functional option for configuring a BehaviorContext.
type ContextOption func(*behaviorContextImpl)

// NewBehaviorContext creates a new BehaviorContext instance with the specified base context.
// The base context can be used for cancellation or passing deadlines to the behavior tree operations.
func NewBehaviorContext(ctx context.Context, options ...ContextOption) BehaviorContext {
	bc := &behaviorContextImpl{
		Ctx:         ctx,
		ContextData: make(map[string]interface{}),
		Blackboard:  NewBlackboard(), // Default blackboard
	}

	// Apply options
	for _, option := range options {
		option(bc)
	}

	return bc
}

// Set stores the value in the behaviorContextImpl for the given key.
// This operation is thread-safe.
func (bc *behaviorContextImpl) Set(key string, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.ContextData[key] = value
}

// Get retrieves the value from the behaviorContextImpl for the given key.
// It returns the value and a boolean indicating whether the key was found.
// This operation is thread-safe.
func (bc *behaviorContextImpl) Get(key string) (value interface{}, ok bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	value, ok = bc.ContextData[key]
	return
}

// Delete removes the value from the behaviorContextImpl for the given key.
// This operation is thread-safe.
func (bc *behaviorContextImpl) Delete(key string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.ContextData, key)
}

// Has checks if the given key exists in the context without retrieving the value.
// This operation is thread-safe and more efficient than Get when only existence needs to be checked.
func (bc *behaviorContextImpl) Has(key string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	_, ok := bc.ContextData[key]
	return ok
}

// GetBlackboard returns the blackboard associated with this context.
func (bc *behaviorContextImpl) GetBlackboard() *Blackboard {
	return bc.Blackboard
}
