package bt

// PrioritySelector represents a behavior tree node that selects the first child that succeeds, and returns Failure if none succeed.
type PrioritySelector struct {
	Composite
	currentRunningIndex int
}

// NewPrioritySelector returns a new PrioritySelector with the given children.
func NewPrioritySelector(children ...Behavior) *PrioritySelector {
	return &PrioritySelector{
		Composite:           Composite{Children: children},
		currentRunningIndex: -1,
	}
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first Success status encountered.
func (s *PrioritySelector) Tick(ctx *BehaviorContext) RunStatus {
	// Iterate over child nodes
	for i := 0; i < len(s.Children); i++ {
		child := s.Children[i]
		// Call Tick method of child node and handle return value
		if child == nil {
			return Failure // handle nil child node
		}
		status := child.Tick(ctx)
		if status != Failure {
			s.currentRunningIndex = -1
			if status == Running {
				s.currentRunningIndex = i
			}
			return status
		}
	}
	// No child node succeeded or is running
	s.currentRunningIndex = -1
	return Failure
}
