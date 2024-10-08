package bt

// BinarySelector is a behavior tree node that conditionally executes one of two child nodes based on a condition.
type BinarySelector struct {
	Condition Behavior
	IfTrue    Behavior
	IfFalse   Behavior
}

// NewBinarySelector creates a new BinarySelector with the given condition, true branch, and false branch.
func NewBinarySelector(condition Behavior, ifTrue Behavior, ifFalse Behavior) *BinarySelector {
	return &BinarySelector{
		Condition: condition,
		IfTrue:    ifTrue,
		IfFalse:   ifFalse,
	}
}

// Tick evaluates the condition and executes the appropriate child node based on the given BehaviorContext.
func (node *BinarySelector) Tick(ctx BehaviorContext) RunStatus {
	if node.Condition.Tick(ctx) == Success {
		return node.IfTrue.Tick(ctx)
	}
	return node.IfFalse.Tick(ctx)
}
