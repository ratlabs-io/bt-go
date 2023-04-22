package bt

// Condition represents a behavior tree node that checks a condition.
type Condition struct {
	// Check is a function that takes a BehaviorContext and returns a boolean value.
	CheckFunc func(ctx *BehaviorContext) bool
}

// NewCondition returns a new Condition with the given check function.
func NewCondition(checkFunc func(ctx *BehaviorContext) bool) *Condition {
	return &Condition{CheckFunc: checkFunc}
}

// Tick evaluates the condition with the given BehaviorContext and returns Success if it's true, otherwise Failure.
func (c *Condition) Tick(ctx *BehaviorContext) RunStatus {
	if c.CheckFunc(ctx) {
		return Success
	}
	return Failure
}

// Conditional represents a behavior tree node that conditionally executes an action.
type Conditional struct {
	Condition *Condition
	Action    Behavior
}

// NewConditional creates a new Conditional with the given condition and action.
func NewConditional(condition *Condition, action Behavior) *Conditional {
	return &Conditional{
		Condition: condition,
		Action:    action,
	}
}

// Tick checks the condition and executes the action with the given BehaviorContext, returning its RunStatus value.
func (c *Conditional) Tick(ctx *BehaviorContext) RunStatus {
	if c.Condition.Tick(ctx) == Success {
		return c.Action.Tick(ctx)
	}
	return Failure
}
