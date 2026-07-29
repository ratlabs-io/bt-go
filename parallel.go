package bt

import (
	"sync"
)

// ParallelPolicy defines how a Parallel node aggregates child results.
type ParallelPolicy int

const (
	// RequireOne succeeds if at least one child succeeds; fails only if all fail.
	RequireOne ParallelPolicy = iota
	// RequireAll succeeds only if all children succeed; fails if any fail.
	RequireAll
	// SuccessOnAll succeeds only if all children succeed; otherwise Running
	// (never Failure — failures are treated as “not yet success”).
	SuccessOnAll
	// SuccessOnOne succeeds if at least one child succeeds; otherwise Running
	// (never Failure).
	SuccessOnOne
)

// String returns the policy name.
func (p ParallelPolicy) String() string {
	switch p {
	case RequireOne:
		return "RequireOne"
	case RequireAll:
		return "RequireAll"
	case SuccessOnAll:
		return "SuccessOnAll"
	case SuccessOnOne:
		return "SuccessOnOne"
	default:
		return "Unknown"
	}
}

// Parallel ticks all children each call and aggregates results by policy.
//
// By default ticks are sequential (deterministic order, same stack) — the usual
// behavior-tree meaning of “parallel”: all children get a chance this frame.
// Use NewConcurrentParallel for one goroutine per child.
//
// Children share the same Env. Under concurrent mode the blackboard is
// thread-safe; other shared mutable state in actions must be synchronized.
//
// Every child is ticked on every Parallel.Tick (no sticky completion memory).
type Parallel struct {
	Composite
	policy     ParallelPolicy
	concurrent bool
}

// NewParallel creates a sequential Parallel node (classic BT parallel).
func NewParallel(policy ParallelPolicy, children ...Behavior) *Parallel {
	return &Parallel{
		Composite:  Composite{Children: children},
		policy:     policy,
		concurrent: false,
	}
}

// NewConcurrentParallel creates a Parallel that ticks each child in its own goroutine.
func NewConcurrentParallel(policy ParallelPolicy, children ...Behavior) *Parallel {
	return &Parallel{
		Composite:  Composite{Children: children},
		policy:     policy,
		concurrent: true,
	}
}

// Policy returns the aggregation policy.
func (p *Parallel) Policy() ParallelPolicy {
	return p.policy
}

// Concurrent reports whether children are ticked in goroutines.
func (p *Parallel) Concurrent() bool {
	return p.concurrent
}

// Tick runs all children (sequentially or concurrently) and returns the policy result.
func (p *Parallel) Tick(env Env) RunStatus {
	n := len(p.Children)
	if n == 0 {
		return Success
	}
	if p.concurrent {
		return p.tickConcurrent(env)
	}
	return p.tickSequential(env)
}

func (p *Parallel) tickSequential(env Env) RunStatus {
	statuses := make([]RunStatus, len(p.Children))
	for i, child := range p.Children {
		if child == nil {
			statuses[i] = Failure
			continue
		}
		statuses[i] = child.Tick(env)
	}
	return evaluateParallelPolicy(p.policy, statuses)
}

func (p *Parallel) tickConcurrent(env Env) RunStatus {
	n := len(p.Children)
	statuses := make([]RunStatus, n)
	var wg sync.WaitGroup

	for i, child := range p.Children {
		if child == nil {
			statuses[i] = Failure
			continue
		}
		wg.Add(1)
		go func(index int, behavior Behavior) {
			defer wg.Done()
			statuses[index] = behavior.Tick(env)
		}(i, child)
	}

	wg.Wait()
	return evaluateParallelPolicy(p.policy, statuses)
}

// Halt aborts every child.
func (p *Parallel) Halt(env Env) {
	for _, child := range p.Children {
		Halt(env, child)
	}
}

func evaluateParallelPolicy(policy ParallelPolicy, statuses []RunStatus) RunStatus {
	successCount := 0
	failureCount := 0

	for _, status := range statuses {
		switch status {
		case Success:
			successCount++
		case Failure:
			failureCount++
		}
	}

	total := len(statuses)

	switch policy {
	case RequireOne:
		if successCount > 0 {
			return Success
		}
		if failureCount == total {
			return Failure
		}
		return Running

	case RequireAll:
		if failureCount > 0 {
			return Failure
		}
		if successCount == total {
			return Success
		}
		return Running

	case SuccessOnAll:
		if successCount == total {
			return Success
		}
		return Running

	case SuccessOnOne:
		if successCount > 0 {
			return Success
		}
		return Running

	default:
		if successCount > 0 {
			return Success
		}
		if failureCount == total {
			return Failure
		}
		return Running
	}
}
