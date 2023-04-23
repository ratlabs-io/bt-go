package bt

// TreeRunner runs a behavior tree with a specified tick rate.
type TreeRunner struct {
	tree Behavior
}

// NewTreeRunner returns a new TreeRunner that runs the given behavior tree with the specified tick rate.
func NewTreeRunner(tree Behavior) *TreeRunner {
	return &TreeRunner{
		tree: tree,
	}
}

// Run runs the behavior tree until the context is done.
func (tr *TreeRunner) Run(ctx *BehaviorContext) {
	for {
		select {
		case <-ctx.Ctx.Done():
			return
		default:
			tr.tree.Tick(ctx)
		}
	}
}
