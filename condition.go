package bt

// Condition represents a behavior tree node that checks a condition.
type Condition struct {
	// Check is a function that takes a BehaviorContext and returns a boolean value.
	Check func(ctx *BehaviorContext) bool
}

// Tick evaluates the condition with the given BehaviorContext and returns Success if it's true, otherwise Failure.
func (c *Condition) Tick(ctx *BehaviorContext) RunStatus {
	if c.Check(ctx) {
		return Success
	}
	return Failure
}

// ConditionBehavior embeds bt.Condition and implements the bt.Behavior interface.
type ConditionBehavior struct {
	*Condition
}
