package bt

import (
	"context"
)

// BehaviorContext is the per-tick environment passed to every node.
//
// Data access (Set/Get/Delete/Has) is backed by a hierarchical Blackboard.
// Context() exposes the underlying context.Context for cancellation and deadlines.
//
// All data methods are safe for concurrent use (via the blackboard).
type BehaviorContext interface {
	// Set stores a value under key on this context's blackboard.
	Set(key string, value interface{})
	// Get retrieves a value by key, walking parent blackboards if needed.
	Get(key string) (value interface{}, ok bool)
	// Delete removes a key from this blackboard only (not parents).
	Delete(key string)
	// Has reports whether key exists on this blackboard or any parent.
	Has(key string) bool
	// GetBlackboard returns the blackboard used for data storage.
	GetBlackboard() *Blackboard
	// Context returns the underlying context.Context for cancellation/deadlines.
	Context() context.Context
}

// behaviorContextImpl is the default BehaviorContext.
type behaviorContextImpl struct {
	ctx        context.Context
	blackboard *Blackboard
}

// ContextOption configures a BehaviorContext at construction time.
type ContextOption func(*behaviorContextImpl)

// WithBlackboard sets the blackboard used by the context.
// If bb is nil, a fresh blackboard is used instead.
func WithBlackboard(bb *Blackboard) ContextOption {
	return func(c *behaviorContextImpl) {
		if bb != nil {
			c.blackboard = bb
		}
	}
}

// NewBehaviorContext creates a BehaviorContext wrapping parent for cancellation.
// A new blackboard is allocated unless WithBlackboard is supplied.
func NewBehaviorContext(parent context.Context, options ...ContextOption) BehaviorContext {
	if parent == nil {
		parent = context.Background()
	}
	bc := &behaviorContextImpl{
		ctx:        parent,
		blackboard: NewBlackboard(),
	}
	for _, option := range options {
		option(bc)
	}
	return bc
}

// Set stores value under key on the blackboard.
func (bc *behaviorContextImpl) Set(key string, value interface{}) {
	bc.blackboard.Set(key, value)
}

// Get retrieves value for key from the blackboard hierarchy.
func (bc *behaviorContextImpl) Get(key string) (value interface{}, ok bool) {
	return bc.blackboard.Get(key)
}

// Delete removes key from this blackboard (not parents).
func (bc *behaviorContextImpl) Delete(key string) {
	bc.blackboard.Delete(key)
}

// Has reports whether key exists in the blackboard hierarchy.
func (bc *behaviorContextImpl) Has(key string) bool {
	return bc.blackboard.Has(key)
}

// GetBlackboard returns the associated blackboard.
func (bc *behaviorContextImpl) GetBlackboard() *Blackboard {
	return bc.blackboard
}

// Context returns the underlying context.Context.
func (bc *behaviorContextImpl) Context() context.Context {
	return bc.ctx
}
