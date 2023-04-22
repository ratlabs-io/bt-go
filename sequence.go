package bt

// SequenceNode represents a behavior tree node that processes its children in sequence until one fails or is running.
type SequenceNode struct {
	Composite
	runningChildIndex int
}

// Tick iterates over the child nodes with the given BehaviorContext and returns Failure if any child fails,
// or Running if any child is running. If all children succeed, returns Success.
func (s *SequenceNode) Tick(ctx *BehaviorContext) RunStatus {
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
	return Success
}
