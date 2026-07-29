package bt

// Action is a leaf node that runs a user-supplied function.
type Action struct {
	runFunc func(ctx BehaviorContext) RunStatus
}

// NewAction creates an Action from runFunc.
// runFunc should return Success, Failure, or Running as appropriate.
// If runFunc is nil, Tick returns Failure.
func NewAction(runFunc func(ctx BehaviorContext) RunStatus) *Action {
	return &Action{runFunc: runFunc}
}

// Tick executes the action function.
func (a *Action) Tick(ctx BehaviorContext) RunStatus {
	if a.runFunc == nil {
		return Failure
	}
	return a.runFunc(ctx)
}
