package bt

import (
	"context"
)

// Env is the per-tick environment passed to every node.
//
// It is intentionally not a context.Context. The two concerns are separate:
//
//   - Context()  — stdlib context.Context for cancellation and deadlines only
//   - Blackboard — mutable agent/world state (hierarchical key-value store)
//
// Do not put agent state in context.WithValue. Use the blackboard (or the
// convenience Set/Get/Delete/Has methods, which write through to it).
//
// Env does not implement context.Context (has-a, not is-a).
type Env interface {
	// Context returns the stdlib context used for cancellation and deadlines.
	// It is never used as a value bag for agent state.
	Context() context.Context

	// Blackboard returns the hierarchical store for agent/world state.
	Blackboard() *Blackboard

	// Set stores a value on the blackboard (convenience for Blackboard().Set).
	Set(key string, value interface{})
	// Get reads a value from the blackboard hierarchy.
	Get(key string) (value interface{}, ok bool)
	// Delete removes a key from this blackboard only (not parents).
	Delete(key string)
	// Has reports whether key exists on this blackboard or any parent.
	Has(key string) bool
}

// env is the default Env implementation.
type env struct {
	ctx        context.Context
	blackboard *Blackboard
}

// EnvOption configures an Env at construction time.
type EnvOption func(*env)

// WithBlackboard sets the blackboard used by the env.
// If bb is nil, a fresh blackboard is used instead.
func WithBlackboard(bb *Blackboard) EnvOption {
	return func(e *env) {
		if bb != nil {
			e.blackboard = bb
		}
	}
}

// NewEnv creates an Env that uses parent for cancellation/deadlines.
// A new blackboard is allocated unless WithBlackboard is supplied.
// If parent is nil, context.Background() is used.
func NewEnv(parent context.Context, options ...EnvOption) Env {
	if parent == nil {
		parent = context.Background()
	}
	e := &env{
		ctx:        parent,
		blackboard: NewBlackboard(),
	}
	for _, option := range options {
		option(e)
	}
	return e
}

// Context returns the stdlib context for cancellation and deadlines.
func (e *env) Context() context.Context {
	return e.ctx
}

// Blackboard returns the agent/world state store.
func (e *env) Blackboard() *Blackboard {
	return e.blackboard
}

// Set stores value under key on the blackboard.
func (e *env) Set(key string, value interface{}) {
	e.blackboard.Set(key, value)
}

// Get retrieves value for key from the blackboard hierarchy.
func (e *env) Get(key string) (value interface{}, ok bool) {
	return e.blackboard.Get(key)
}

// Delete removes key from this blackboard (not parents).
func (e *env) Delete(key string) {
	e.blackboard.Delete(key)
}

// Has reports whether key exists in the blackboard hierarchy.
func (e *env) Has(key string) bool {
	return e.blackboard.Has(key)
}
