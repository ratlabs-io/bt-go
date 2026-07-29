package bt

// Selector ticks children left-to-right until one succeeds or is still running.
//
// Semantics (reactive / restart-from-start each tick — classic priority order):
//   - Success from any child → Success
//   - Running from any child → Running
//   - All Failure → Failure
//
// Earlier children have higher priority: every Tick re-evaluates from the first
// child, so a higher-priority branch can preempt a lower one that was Running.
// For stick-to-running-child semantics (no preemption), use NewMemorySelector.
//
// PrioritySelector is an alias for Selector with identical semantics.
type Selector struct {
	Composite
}

// NewSelector creates a Selector with the given children, in priority order.
func NewSelector(children ...Behavior) *Selector {
	return &Selector{
		Composite: Composite{Children: children},
	}
}

// Tick tries each child until one does not fail. See type docs for status rules.
func (s *Selector) Tick(env Env) RunStatus {
	for _, child := range s.Children {
		if child == nil {
			return Failure
		}
		status := child.Tick(env)
		if status != Failure {
			return status
		}
	}
	return Failure
}
