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

// Parallel ticks all children concurrently (one goroutine each) and aggregates
// results according to policy.
//
// Children share the same Env. Because the blackboard is
// thread-safe, concurrent Set/Get is safe; actions that mutate other shared
// state must synchronize themselves.
//
// Every child is ticked on every Parallel.Tick — there is no “skip completed
// children” memory. That keeps policies like SuccessOnAll correct when a child
// can recover from Failure on a later tick.
type Parallel struct {
	Composite
	policy ParallelPolicy
}

// NewParallel creates a Parallel node with the given policy and children.
func NewParallel(policy ParallelPolicy, children ...Behavior) *Parallel {
	return &Parallel{
		Composite: Composite{Children: children},
		policy:    policy,
	}
}

// Policy returns the aggregation policy.
func (p *Parallel) Policy() ParallelPolicy {
	return p.policy
}

// Tick runs all children concurrently and returns the policy result.
func (p *Parallel) Tick(env Env) RunStatus {
	n := len(p.Children)
	if n == 0 {
		return Success
	}

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
			// Share the same Env so blackboard writes stay coherent.
			statuses[index] = behavior.Tick(env)
		}(i, child)
	}

	wg.Wait()
	return evaluateParallelPolicy(p.policy, statuses)
}

func evaluateParallelPolicy(policy ParallelPolicy, statuses []RunStatus) RunStatus {
	successCount := 0
	failureCount := 0
	runningCount := 0

	for _, status := range statuses {
		switch status {
		case Success:
			successCount++
		case Failure:
			failureCount++
		case Running:
			runningCount++
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
		// Unknown policy: same as RequireOne.
		if successCount > 0 {
			return Success
		}
		if failureCount == total {
			return Failure
		}
		return Running
	}
}
