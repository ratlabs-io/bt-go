package bt

import (
	"time"
)

// TreeRunner runs a behavior tree with a specified tick rate.
type TreeRunner struct {
	tree      Behavior
	tickRate  time.Duration
	onSuccess func()
	onFailure func()
	onRunning func()
}

// RunnerOption is a functional option for configuring a TreeRunner.
type RunnerOption func(*TreeRunner)

// WithTickRate sets the tick rate for the runner.
func WithTickRate(rate time.Duration) RunnerOption {
	return func(tr *TreeRunner) {
		tr.tickRate = rate
	}
}

// WithCallbacks sets callback functions for the runner.
func WithCallbacks(onSuccess, onFailure, onRunning func()) RunnerOption {
	return func(tr *TreeRunner) {
		tr.onSuccess = onSuccess
		tr.onFailure = onFailure
		tr.onRunning = onRunning
	}
}

// NewTreeRunner returns a new TreeRunner that runs the given behavior tree with the specified options.
func NewTreeRunner(tree Behavior, options ...RunnerOption) *TreeRunner {
	tr := &TreeRunner{
		tree:      tree,
		tickRate:  time.Millisecond * 100, // Default tick rate: 100ms (10Hz)
		onSuccess: func() {},              // Default empty callbacks
		onFailure: func() {},
		onRunning: func() {},
	}

	// Apply options
	for _, option := range options {
		option(tr)
	}

	return tr
}

// Run runs the behavior tree until the context is done.
func (tr *TreeRunner) Run(ctx BehaviorContext) {
	ticker := time.NewTicker(tr.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.(*behaviorContextImpl).Ctx.Done():
			return
		case <-ticker.C:
			status := tr.tree.Tick(ctx)

			switch status {
			case Success:
				tr.onSuccess()
			case Failure:
				tr.onFailure()
			case Running:
				tr.onRunning()
			}
		}
	}
}

// RunOnce runs the behavior tree once and returns its status.
func (tr *TreeRunner) RunOnce(ctx BehaviorContext) RunStatus {
	status := tr.tree.Tick(ctx)

	switch status {
	case Success:
		tr.onSuccess()
	case Failure:
		tr.onFailure()
	case Running:
		tr.onRunning()
	}

	return status
}
