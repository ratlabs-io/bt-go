package bt

// Selector represents a behavior tree node that selects the first child that succeeds or is running.
type Selector struct {
	Composite
	runningChildIndex int
}

// NewSelector returns a new Selector with the given children.
func NewSelector(children ...Behavior) *Selector {
	return &Selector{
		Composite:         Composite{Children: children},
		runningChildIndex: -1,
	}
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first non-Failure status encountered.
func (s *Selector) Tick(ctx *BehaviorContext) RunStatus {
	for i, child := range s.Children {
		switch status := child.Tick(ctx); status {
		case Success:
			s.runningChildIndex = -1
			return Success
		case Running:
			s.runningChildIndex = i
			return Running
		}
	}
	s.runningChildIndex = -1
	return Failure
}
