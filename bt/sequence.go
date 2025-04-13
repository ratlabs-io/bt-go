package bt

// Sequence represents a composite node in a behavior tree that executes its child nodes in order until one fails or is still running.
// It is used to ensure a series of behaviors are performed sequentially, stopping at the first sign of failure or ongoing execution.
type Sequence struct {
	Composite             // Composite embeds the base structure for holding child behaviors.
	runningChildIndex int // runningChildIndex tracks the index of the currently running child, if any.
}

// NewSequence creates a new Sequence node with the provided child behaviors.
// The children will be executed in the order they are provided.
func NewSequence(children ...Behavior) *Sequence {
	return &Sequence{
		Composite:         Composite{Children: children},
		runningChildIndex: -1,
	}
}

// Tick executes each child node in sequence using the given BehaviorContext.
// It returns Failure if any child fails, Running if any child is still executing, or Success if all children succeed.
// The state of the running child is tracked to allow resuming from the correct point on subsequent ticks.
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
