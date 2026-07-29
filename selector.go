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
// When preemption occurs, the abandoned child is Halted if it implements Haltable.
// For stick-to-running-child semantics (no preemption), use NewMemorySelector.
type Selector struct {
	Composite
	// lastRunning tracks the child that returned Running on a prior tick (for Halt).
	lastRunning int
}

// NewSelector creates a Selector with the given children, in priority order.
func NewSelector(children ...Behavior) *Selector {
	return &Selector{
		Composite:   Composite{Children: children},
		lastRunning: -1,
	}
}

// Tick tries each child until one does not fail. See type docs for status rules.
func (s *Selector) Tick(env Env) RunStatus {
	chosen := -1
	status := Failure

	for i, child := range s.Children {
		if child == nil {
			s.abortRunning(env, -1)
			return Failure
		}
		status = child.Tick(env)
		if status != Failure {
			chosen = i
			break
		}
	}

	if status == Failure {
		s.abortRunning(env, -1)
		return Failure
	}

	// Preempt: higher-priority child took over (or a different branch).
	if s.lastRunning >= 0 && s.lastRunning != chosen {
		s.abortRunning(env, chosen)
	}

	if status == Running {
		s.lastRunning = chosen
	} else {
		s.lastRunning = -1
	}
	return status
}

// Halt aborts the last running child and clears tracking.
func (s *Selector) Halt(env Env) {
	s.abortRunning(env, -1)
}

func (s *Selector) abortRunning(env Env, keep int) {
	if s.lastRunning >= 0 && s.lastRunning != keep && s.lastRunning < len(s.Children) {
		Halt(env, s.Children[s.lastRunning])
	}
	if keep < 0 {
		s.lastRunning = -1
	}
}
