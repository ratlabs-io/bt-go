package bt

// Selector represents a behavior tree node that selects the first child that succeeds or is running.
type Selector struct {
	Children          []Behavior
	runningChildIndex int
}

// GetChildren returns the children of the Selector.
func (s *Selector) GetChildren() []Behavior {
	return s.Children
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
