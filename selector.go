package bt

// Selector represents a behavior tree node that selects the first child that succeeds or is running.
type Selector struct {
	Composite
	runningChildIndex int
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first non-Failure status encountered.
func (s *Selector) Tick(ctx *BehaviorContext) RunStatus {
	for i, child := range s.Children {
		status := child.Tick(ctx)
		if status != Failure {
			if status == Running {
				s.runningChildIndex = i
			} else {
				s.runningChildIndex = -1
			}
			return status
		}
	}
	return Failure
}
