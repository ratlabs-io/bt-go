package bt

// Action represents a behavior tree node that performs an action.
type Action struct {
	runFunc func(ctx BehaviorContext) RunStatus
}

// NewAction creates a new Action with the given function.
func NewAction(runFunc func(ctx BehaviorContext) RunStatus) *Action {
	return &Action{
		runFunc: runFunc,
	}
}

// Tick executes the action's function with the given BehaviorContext and returns its RunStatus.
func (a *Action) Tick(ctx BehaviorContext) RunStatus {
	return a.runFunc(ctx)
}
