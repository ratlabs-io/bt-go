package bt

// Action represents a behavior tree node that performs an action.
type Action struct {
	// Action is a function that takes a BehaviorContext and returns a NodeStatus.
	Action func(ctx *BehaviorContext) RunStatus
}

// Tick calls the action with the given context and returns its status.
func (a *Action) Tick(ctx *BehaviorContext) RunStatus {
	return a.Action(ctx)
}
