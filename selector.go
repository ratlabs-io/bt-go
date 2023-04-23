package bt

// Selector represents a behavior tree node that selects the first child that succeeds or is running.
type Selector struct {
	Composite
	currentRunningIndex int
}

// NewSelector returns a new Selector with the given children.
func NewSelector(children ...Behavior) *Selector {
	return &Selector{
		Composite:           Composite{Children: children},
		currentRunningIndex: -1,
	}
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first non-Failure status encountered.
func (s *Selector) Tick(ctx *BehaviorContext) RunStatus {
	// Iterate over child nodes
	for i, child := range s.Children {
		// Call Tick method of child node and handle return value
		if child == nil {
			return Failure // handle nil child node
		}
		status := child.Tick(ctx)
		if status == Success {
			s.currentRunningIndex = -1
			return Success
		} else if status == Running {
			s.currentRunningIndex = i
			return Running
		}
	}
	// No child node succeeded or is running
	s.currentRunningIndex = -1
	return Failure
}
