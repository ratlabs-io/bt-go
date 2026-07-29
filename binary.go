package bt

// BinarySelector chooses between two branches based on a condition behavior.
//
// If Condition returns Success, IfTrue is ticked; otherwise IfFalse is ticked.
// Condition statuses other than Success (including Running) select IfFalse.
// For a boolean leaf, pass a *Condition.
//
// When the selected branch changes while the previous branch was Running, the
// abandoned branch is Halted. Halt propagates to the last Running branch.
type BinarySelector struct {
	Condition Behavior
	IfTrue    Behavior
	IfFalse   Behavior
	// lastBranch: -1 none, 0 IfTrue, 1 IfFalse (only set when that branch returned Running).
	lastBranch int
}

// NewBinarySelector creates a BinarySelector.
func NewBinarySelector(condition, ifTrue, ifFalse Behavior) *BinarySelector {
	return &BinarySelector{
		Condition:  condition,
		IfTrue:     ifTrue,
		IfFalse:    ifFalse,
		lastBranch: -1,
	}
}

// Tick evaluates Condition and runs the matching branch.
func (node *BinarySelector) Tick(env Env) RunStatus {
	chosen := 1 // IfFalse
	if node.Condition != nil && node.Condition.Tick(env) == Success {
		chosen = 0
	}

	var branch Behavior
	if chosen == 0 {
		branch = node.IfTrue
	} else {
		branch = node.IfFalse
	}
	if branch == nil {
		node.abortBranch(env, -1)
		return Failure
	}

	status := branch.Tick(env)

	if node.lastBranch >= 0 && node.lastBranch != chosen {
		node.abortBranch(env, chosen)
	}

	if status == Running {
		node.lastBranch = chosen
	} else {
		node.lastBranch = -1
	}
	return status
}

// Halt aborts the last running branch.
func (node *BinarySelector) Halt(env Env) {
	node.abortBranch(env, -1)
}

func (node *BinarySelector) abortBranch(env Env, keep int) {
	if node.lastBranch >= 0 && node.lastBranch != keep {
		switch node.lastBranch {
		case 0:
			Halt(env, node.IfTrue)
		case 1:
			Halt(env, node.IfFalse)
		}
	}
	if keep < 0 {
		node.lastBranch = -1
	}
}
