package bt

// Action represents a leaf node in a behavior tree that executes a specific task or operation.
// It encapsulates a function that is called when the node is ticked, determining the outcome of the behavior.
type Action struct {
	runFunc func(ctx BehaviorContext) RunStatus // runFunc is the function executed when the action is ticked.
}

// NewAction creates a new Action instance with the provided function.
// The function should implement the logic of the action and return the appropriate RunStatus.
func NewAction(runFunc func(ctx BehaviorContext) RunStatus) *Action {
	return &Action{
		runFunc: runFunc,
	}
}

// Tick executes the action's function with the given BehaviorContext and returns its RunStatus.
// This method satisfies the Behavior interface, allowing Action to be used as a node in the behavior tree.
func (a *Action) Tick(ctx BehaviorContext) RunStatus {
	return a.runFunc(ctx)
}
