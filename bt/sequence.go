package bt

// Sequence ticks children left-to-right until one fails or is still running.
//
// Semantics (reactive / restart-from-start each tick):
//   - Failure from any child → Failure
//   - Running from any child → Running
//   - All Success → Success
//
// Sequence does not remember which child was running; each Tick starts at the
// first child. For resume-from-running semantics, use NewMemorySequence.
type Sequence struct {
	Composite
}

// NewSequence creates a Sequence with the given children, in order.
func NewSequence(children ...Behavior) *Sequence {
	return &Sequence{
		Composite: Composite{Children: children},
	}
}

// Tick executes children in order. See type docs for status rules.
func (s *Sequence) Tick(ctx BehaviorContext) RunStatus {
	for _, child := range s.Children {
		if child == nil {
			return Failure
		}
		status := child.Tick(ctx)
		if status != Success {
			return status
		}
	}
	return Success
}
