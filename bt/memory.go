package bt

// MemorySequence is a Sequence that remembers which child was Running.
//
// Unlike the reactive Sequence (NewSequence), once a child returns Running the
// next Tick resumes at that child and does not re-tick earlier siblings that
// already succeeded in this run.
//
// Memory is cleared when the sequence returns Success or Failure (including
// nil-child failure). Call Reset to clear memory without ticking.
type MemorySequence struct {
	Composite
	// runningIndex is the child to resume from. 0 means start (or resume) at first.
	runningIndex int
}

// NewMemorySequence creates a memory Sequence with the given children.
func NewMemorySequence(children ...Behavior) *MemorySequence {
	return &MemorySequence{
		Composite:    Composite{Children: children},
		runningIndex: 0,
	}
}

// Reset clears resume state so the next Tick starts at the first child.
func (s *MemorySequence) Reset() {
	s.runningIndex = 0
}

// RunningIndex returns the child index that will be resumed on the next Tick
// (0 when idle / after a terminal status).
func (s *MemorySequence) RunningIndex() int {
	return s.runningIndex
}

// Tick executes children from the remembered index. See type docs.
func (s *MemorySequence) Tick(ctx BehaviorContext) RunStatus {
	if s.runningIndex < 0 || s.runningIndex > len(s.Children) {
		s.runningIndex = 0
	}

	for i := s.runningIndex; i < len(s.Children); i++ {
		child := s.Children[i]
		if child == nil {
			s.runningIndex = 0
			return Failure
		}
		status := child.Tick(ctx)
		switch status {
		case Running:
			s.runningIndex = i
			return Running
		case Failure:
			s.runningIndex = 0
			return Failure
		case Success:
			// Advance past this child; next loop iteration continues.
			s.runningIndex = i + 1
		}
	}

	s.runningIndex = 0
	return Success
}

// MemorySelector is a Selector that remembers which child was Running.
//
// Unlike the reactive Selector (NewSelector), once a child returns Running the
// next Tick resumes at that child and does not re-evaluate higher-priority
// (earlier) siblings until the remembered child finishes.
//
// If the remembered child fails, evaluation continues with later siblings only
// (earlier ones are not re-tried until the whole selector returns Failure and
// memory is cleared). Call Reset to clear memory without ticking.
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
func (s *MemorySelector) Reset() {
	s.runningIndex = -1
}

// RunningIndex returns the remembered child index, or -1 when idle.
func (s *MemorySelector) RunningIndex() int {
	return s.runningIndex
}

// Tick tries children from the remembered index. See type docs.
func (s *MemorySelector) Tick(ctx BehaviorContext) RunStatus {
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
		status := child.Tick(ctx)
		switch status {
		case Running:
			s.runningIndex = i
			return Running
		case Success:
			s.runningIndex = -1
			return Success
		case Failure:
			// Try the next sibling; clear sticky index so we don't re-stick on fail.
			s.runningIndex = -1
			// But if we continue the loop, we should not re-check earlier children.
			// Keep start progressing by leaving runningIndex idle and continuing i+1.
		}
	}

	s.runningIndex = -1
	return Failure
}
