package bt

// Condition is a leaf that maps a boolean check to Success or Failure.
// It never returns Running.
type Condition struct {
	checkFunc func(ctx BehaviorContext) bool
}

// NewCondition creates a Condition from checkFunc.
// If checkFunc is nil, Tick returns Failure.
func NewCondition(checkFunc func(ctx BehaviorContext) bool) *Condition {
	return &Condition{checkFunc: checkFunc}
}

// Tick returns Success when the check is true, otherwise Failure.
func (c *Condition) Tick(ctx BehaviorContext) RunStatus {
	if c.checkFunc == nil {
		return Failure
	}
	if c.checkFunc(ctx) {
		return Success
	}
	return Failure
}

// Conditional runs Action only when Condition succeeds.
// If the condition fails, Conditional returns Failure without ticking Action.
// Condition must be non-nil; a nil Action is treated as Failure when selected.
type Conditional struct {
	Condition *Condition
	Action    Behavior
}

// NewConditional creates a Conditional with the given condition and action.
func NewConditional(condition *Condition, action Behavior) *Conditional {
	return &Conditional{
		Condition: condition,
		Action:    action,
	}
}

// Tick checks the condition, then optionally runs the action.
func (c *Conditional) Tick(ctx BehaviorContext) RunStatus {
	if c.Condition == nil || c.Condition.Tick(ctx) != Success {
		return Failure
	}
	if c.Action == nil {
		return Failure
	}
	return c.Action.Tick(ctx)
}
