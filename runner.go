package bt

import (
	"time"
)

// TreeRunner runs a behavior tree with a specified tick rate.
type TreeRunner struct {
	tree     Behavior
	interval time.Duration
}

// NewTreeRunner returns a new TreeRunner that runs the given behavior tree with the specified tick rate.
func NewTreeRunner(tree Behavior, tickRate time.Duration) *TreeRunner {
	return &TreeRunner{
		tree:     tree,
		interval: tickRate,
	}
}

// Run runs the behavior tree until the context is done.
// Run runs the behavior tree with the given context, executing the tree at the specified interval.
// If the interval is 0, the tree is run continuously without waiting.
func (tr *TreeRunner) Run(ctx *BehaviorContext) {
	if tr.interval == 0 {
		for {
			select {
			case <-ctx.Ctx.Done():
				return
			default:
				tr.tree.Tick(ctx)
			}
		}
	}

	ticker := time.NewTicker(tr.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Ctx.Done():
			return
		case <-ticker.C:
			tr.tree.Tick(ctx)
		}
	}
}
