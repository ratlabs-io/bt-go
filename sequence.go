package bt

// Sequence ticks children left-to-right until one fails or is still running.
//
// Semantics (reactive / restart-from-start each tick):
//   - Failure from any child → Failure
//   - Running from any child → Running
//   - All Success → Success
//
// Sequence does not remember progress for control flow; each Tick starts at the
// first child. It does track the last Running child so that if an earlier sibling
// later fails (or the sequence is Halted), the abandoned child receives Halt.
// For resume-from-running control flow, use NewMemorySequence.
type Sequence struct {
	Composite
	lastRunning int
}

// NewSequence creates a Sequence with the given children, in order.
func NewSequence(children ...Behavior) *Sequence {
	return &Sequence{
		Composite:   Composite{Children: children},
		lastRunning: -1,
	}
}

// Tick executes children in order. See type docs for status rules.
func (s *Sequence) Tick(env Env) RunStatus {
	for i, child := range s.Children {
		if child == nil {
			s.abortRunning(env)
			return Failure
		}
		status := child.Tick(env)
		if status == Running {
			if s.lastRunning >= 0 && s.lastRunning != i {
				// Should be rare in a reactive sequence; still safe.
				Halt(env, s.Children[s.lastRunning])
			}
			s.lastRunning = i
			return Running
		}
		if status == Failure {
			// Earlier sibling failed after a later one had been running.
			if s.lastRunning > i {
				Halt(env, s.Children[s.lastRunning])
			}
			s.lastRunning = -1
			return Failure
		}
	}
	s.lastRunning = -1
	return Success
}

// Halt aborts the last running child.
func (s *Sequence) Halt(env Env) {
	s.abortRunning(env)
}

func (s *Sequence) abortRunning(env Env) {
	if s.lastRunning >= 0 && s.lastRunning < len(s.Children) {
		Halt(env, s.Children[s.lastRunning])
	}
	s.lastRunning = -1
}
