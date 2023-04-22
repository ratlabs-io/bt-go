package bt

// PrioritySelector represents a behavior tree node that selects the first child that succeeds, and returns Failure if none succeed.
type PrioritySelector struct {
	Composite
	runningChildIndex int
}

// Tick iterates over the child nodes with the given BehaviorContext and returns the first Success status encountered.
func (ps *PrioritySelector) Tick(ctx *BehaviorContext) RunStatus {
	for i, child := range ps.Children {
		status := child.Tick(ctx)
		if status != Failure {
			if status == Running {
				ps.runningChildIndex = i
			} else {
				ps.runningChildIndex = -1
			}
			return status
		}
	}
	return Failure
}

// Abort aborts the currently running child node, if any.
func (ps *PrioritySelector) Abort(ctx *BehaviorContext) {
	if ps.runningChildIndex >= 0 && ps.runningChildIndex < len(ps.Children) {
		ps.Children[ps.runningChildIndex].Abort(ctx)
	}
}
