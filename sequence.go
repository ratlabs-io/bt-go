package bt

// Sequence represents a behavior tree node that processes its children in sequence until one fails or is running.
type Sequence struct {
	Composite
	runningChildIndex int
}

// NewSequence returns a new Sequence with the given children.
func NewSequence(children ...Behavior) *Sequence {
	return &Sequence{
		Composite:         Composite{Children: children},
		runningChildIndex: -1,
	}
}

// Tick iterates over the child nodes with the given BehaviorContext and returns Failure if any child fails,
// or Running if any child is running. If all children succeed, returns Success.
func (s *Sequence) Tick(ctx BehaviorContext) RunStatus {
	for i, child := range s.Children {
		status := child.Tick(ctx)
		if status != Success {
			if status == Running {
				s.runningChildIndex = i
			} else {
				s.runningChildIndex = -1
			}
			return status
		}
	}
	s.runningChildIndex = -1
	return Success
}
