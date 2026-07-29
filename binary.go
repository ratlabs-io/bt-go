package bt

// BinarySelector chooses between two branches based on a condition behavior.
//
// If Condition returns Success, IfTrue is ticked; otherwise IfFalse is ticked.
// Condition statuses other than Success (including Running) select IfFalse.
// For a boolean leaf, pass a *Condition.
type BinarySelector struct {
	Condition Behavior
	IfTrue    Behavior
	IfFalse   Behavior
}

// NewBinarySelector creates a BinarySelector.
func NewBinarySelector(condition, ifTrue, ifFalse Behavior) *BinarySelector {
	return &BinarySelector{
		Condition: condition,
		IfTrue:    ifTrue,
		IfFalse:   ifFalse,
	}
}

// Tick evaluates Condition and runs the matching branch.
func (node *BinarySelector) Tick(env Env) RunStatus {
	if node.Condition != nil && node.Condition.Tick(env) == Success {
		if node.IfTrue == nil {
			return Failure
		}
		return node.IfTrue.Tick(env)
	}
	if node.IfFalse == nil {
		return Failure
	}
	return node.IfFalse.Tick(env)
}
