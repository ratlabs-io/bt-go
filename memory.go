package bt

// MemorySequence is a Sequence that remembers which child was Running.
//
// Unlike the reactive Sequence (NewSequence), once a child returns Running the
// next Tick resumes at that child and does not re-tick earlier siblings that
// already succeeded in this run.
//
// Memory is cleared when the sequence returns Success or Failure (including
// nil-child failure). Call Reset to clear memory without ticking; Halt aborts
// the current child and resets. Idle runningIndex is -1 (Halt is a no-op).
type MemorySequence struct {
	Composite
	// runningIndex is the child to resume from, or -1 when idle.
	runningIndex int
}

// NewMemorySequence creates a memory Sequence with the given children.
func NewMemorySequence(children ...Behavior) *MemorySequence {
	return &MemorySequence{
		Composite:    Composite{Children: children},
		runningIndex: -1,
	}
}

// Reset clears resume state so the next Tick starts at the first child.
// It does not Halt the current child; call Halt when aborting mid-run.
func (s *MemorySequence) Reset() {
	s.runningIndex = -1
}

// RunningIndex returns the child index that will be resumed on the next Tick,
// or -1 when idle / after a terminal status.
func (s *MemorySequence) RunningIndex() int {
	return s.runningIndex
}

// Tick executes children from the remembered index. See type docs.
func (s *MemorySequence) Tick(env Env) RunStatus {
	start := 0
	if s.runningIndex >= 0 {
		if s.runningIndex >= len(s.Children) {
			s.runningIndex = -1
		} else {
			start = s.runningIndex
		}
	}

	for i := start; i < len(s.Children); i++ {
		child := s.Children[i]
		if child == nil {
			s.runningIndex = -1
			return Failure
		}
		status := child.Tick(env)
		switch status {
		case Running:
			s.runningIndex = i
			return Running
		case Failure:
			s.runningIndex = -1
			return Failure
		case Success:
			// Continue; park only if we return Running later this tick.
			s.runningIndex = -1
		}
	}

	s.runningIndex = -1
	return Success
}

// Halt aborts the child at the resume index (if any) and resets memory.
func (s *MemorySequence) Halt(env Env) {
	if s.runningIndex >= 0 && s.runningIndex < len(s.Children) {
		Halt(env, s.Children[s.runningIndex])
	}
	s.runningIndex = -1
}

// MemorySelector is a Selector that remembers which child was Running.
//
// Unlike the reactive Selector (NewSelector), once a child returns Running the
// next Tick resumes at that child and does not re-evaluate higher-priority
// (earlier) siblings until the remembered child finishes.
//
// If the remembered child fails, evaluation continues with later siblings only
// (earlier ones are not re-tried until the whole selector returns Failure and
// memory is cleared). Call Reset to clear memory without ticking; Halt aborts.
type MemorySelector struct {
	Composite
	// runningIndex is -1 when idle (start from first child), or the index to resume.
	runningIndex int
}

// NewMemorySelector creates a memory Selector with the given children.
func NewMemorySelector(children ...Behavior) *MemorySelector {
	return &MemorySelector{
		Composite:    Composite{Children: children},
		runningIndex: -1,
	}
}

// Reset clears resume state so the next Tick starts at the first child.
// It does not Halt the current child; call Halt when aborting mid-run.
func (s *MemorySelector) Reset() {
	s.runningIndex = -1
}

// RunningIndex returns the remembered child index, or -1 when idle.
func (s *MemorySelector) RunningIndex() int {
	return s.runningIndex
}

// Tick tries children from the remembered index. See type docs.
func (s *MemorySelector) Tick(env Env) RunStatus {
	start := 0
	if s.runningIndex >= 0 {
		if s.runningIndex >= len(s.Children) {
			s.runningIndex = -1
		} else {
			start = s.runningIndex
		}
	}

	for i := start; i < len(s.Children); i++ {
		child := s.Children[i]
		if child == nil {
			s.runningIndex = -1
			return Failure
		}
		status := child.Tick(env)
		switch status {
		case Running:
			s.runningIndex = i
			return Running
		case Success:
			s.runningIndex = -1
			return Success
		case Failure:
			// Child finished with Failure — no Halt. Try next sibling.
			s.runningIndex = -1
		}
	}

	s.runningIndex = -1
	return Failure
}

// Halt aborts the remembered child (if any) and clears memory.
func (s *MemorySelector) Halt(env Env) {
	if s.runningIndex >= 0 && s.runningIndex < len(s.Children) {
		Halt(env, s.Children[s.runningIndex])
	}
	s.runningIndex = -1
}
