package bt

// PrioritySelector represents a behavior tree node that selects the first child that succeeds, and returns Failure if none succeed.
type PrioritySelector struct {
	Composite
	runningChildIndex int
}

// NewPrioritySelector returns a new PrioritySelector with the given children.
func NewPrioritySelector(children ...Behavior) *PrioritySelector {
	return &PrioritySelector{
		Composite:         Composite{Children: children},
		runningChildIndex: -1,
	}
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
