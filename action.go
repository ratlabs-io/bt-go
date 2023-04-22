package bt

// Action represents a behavior tree node that performs an action.
type Action struct {
	// Action is a function that takes a BehaviorContext and returns a NodeStatus.
	Action func(ctx *BehaviorContext) RunStatus
	// AbortAction is a function that aborts the action's execution. It takes a BehaviorContext as an argument.
	AbortAction func(ctx *BehaviorContext)
}

// Tick calls the action with the given context and returns its status.
func (a *Action) Tick(ctx *BehaviorContext) RunStatus {
	return a.Action(ctx)
}

// Abort calls the abort action with the given context.
func (a *Action) Abort(ctx *BehaviorContext) {
	a.AbortAction(ctx)
}
