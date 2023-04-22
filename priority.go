package bt

// PrioritySelector represents a behavior tree node that selects the first child that succeeds, and returns Failure if none succeed.
type PrioritySelector struct {
	Children          []Behavior
	runningChildIndex int
}

// GetChildren returns the children of the PrioritySelector.
func (s *PrioritySelector) GetChildren() []Behavior {
	return s.Children
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first Success status encountered.
func (s *PrioritySelector) Tick(ctx *BehaviorContext) RunStatus {
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
