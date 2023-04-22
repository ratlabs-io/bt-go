package bt

// BinarySelector is a behavior tree node that conditionally executes one of two child nodes based on a condition.
type BinarySelector struct {
	Condition Behavior
	IfTrue    Behavior
	IfFalse   Behavior
}

// Tick evaluates the condition and executes the appropriate child node based on the given BehaviorContext.
func (node *BinarySelector) Tick(ctx *BehaviorContext) RunStatus {
	if node.Condition.Tick(ctx) == Success {
		return node.IfTrue.Tick(ctx)
	}
	return node.IfFalse.Tick(ctx)
}
