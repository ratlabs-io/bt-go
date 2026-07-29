package bt

import (
	"time"
)

// TreeRunner ticks a behavior tree on a fixed interval until cancelled.
type TreeRunner struct {
	tree      Behavior
	tickRate  time.Duration
	onSuccess func()
	onFailure func()
	onRunning func()
}

// RunnerOption configures a TreeRunner.
type RunnerOption func(*TreeRunner)

// WithTickRate sets how often the tree is ticked. Default is 100ms (10 Hz).
func WithTickRate(rate time.Duration) RunnerOption {
	return func(tr *TreeRunner) {
		tr.tickRate = rate
	}
}

// WithCallbacks registers hooks invoked after each tick for the returned status.
// Nil callbacks are treated as no-ops.
func WithCallbacks(onSuccess, onFailure, onRunning func()) RunnerOption {
	return func(tr *TreeRunner) {
		if onSuccess != nil {
			tr.onSuccess = onSuccess
		}
		if onFailure != nil {
			tr.onFailure = onFailure
		}
		if onRunning != nil {
			tr.onRunning = onRunning
		}
	}
}

// NewTreeRunner returns a runner for tree with optional configuration.
func NewTreeRunner(tree Behavior, options ...RunnerOption) *TreeRunner {
	tr := &TreeRunner{
		tree:      tree,
		tickRate:  time.Millisecond * 100,
		onSuccess: func() {},
		onFailure: func() {},
		onRunning: func() {},
	}
	for _, option := range options {
		option(tr)
	}
	return tr
}

// Run ticks the tree at the configured rate until ctx is cancelled.
// It blocks until Context().Done() is closed.
func (tr *TreeRunner) Run(ctx BehaviorContext) {
	ticker := time.NewTicker(tr.tickRate)
	defer ticker.Stop()

	done := ctx.Context().Done()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			tr.dispatch(tr.tree.Tick(ctx))
		}
	}
}

// RunOnce ticks the tree once, fires the matching callback, and returns the status.
func (tr *TreeRunner) RunOnce(ctx BehaviorContext) RunStatus {
	status := tr.tree.Tick(ctx)
	tr.dispatch(status)
	return status
}

func (tr *TreeRunner) dispatch(status RunStatus) {
	switch status {
	case Success:
		tr.onSuccess()
	case Failure:
		tr.onFailure()
	case Running:
		tr.onRunning()
	}
}
