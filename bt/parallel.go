package bt

import (
	"sync"
)

// ParallelPolicy defines how a Parallel node determines its return status based on child results.
type ParallelPolicy int

const (
	// RequireOne means the Parallel succeeds if at least one child succeeds, fails if all fail.
	RequireOne ParallelPolicy = iota
	// RequireAll means the Parallel succeeds only if all children succeed, fails if any fail.
	RequireAll
	// SuccessOnAll means the Parallel succeeds only if all children succeed, returns Running otherwise.
	SuccessOnAll
	// SuccessOnOne means the Parallel succeeds if at least one child succeeds, returns Running otherwise.
	SuccessOnOne
)

// Parallel is a composite node that runs all its children concurrently.
type Parallel struct {
	Composite
	policy      ParallelPolicy
	childStatus []RunStatus
}

// NewParallel creates a new Parallel node with the given children and success policy.
func NewParallel(policy ParallelPolicy, children ...Behavior) *Parallel {
	childStatus := make([]RunStatus, len(children))
	// Initialize all statuses to Running
	for i := range childStatus {
		childStatus[i] = Running
	}

	return &Parallel{
		Composite:   Composite{Children: children},
		policy:      policy,
		childStatus: childStatus,
	}
}

// Tick runs all child nodes concurrently and aggregates their results according to the policy.
func (p *Parallel) Tick(ctx BehaviorContext) RunStatus {
	// Reset status if length mismatch (e.g. children added/removed)
	if len(p.childStatus) != len(p.Children) {
		p.childStatus = make([]RunStatus, len(p.Children))
		for i := range p.childStatus {
			p.childStatus[i] = Running
		}
	}

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	var mu sync.Mutex // To protect access to childStatus

	// Start a goroutine for each child
	for i, child := range p.Children {
		if child == nil {
			p.childStatus[i] = Failure
			continue
		}

		// Only start a goroutine for children still running
		if p.childStatus[i] == Running {
			wg.Add(1)
			go func(index int, behavior Behavior) {
				defer wg.Done()

				// Create a separate context for each child to avoid race conditions
				// We only use the blackboard from the main context
				childCtx := NewBehaviorContext(ctx.(*behaviorContextImpl).Ctx,
					WithBlackboard(ctx.(*behaviorContextImpl).Blackboard))

				status := behavior.Tick(childCtx)

				mu.Lock()
				p.childStatus[index] = status
				mu.Unlock()
			}(i, child)
		}
	}

	// Wait for all children to complete execution
	wg.Wait()

	// Determine the result based on the policy
	return p.evaluatePolicy()
}

// evaluatePolicy determines the final status based on child statuses and policy.
func (p *Parallel) evaluatePolicy() RunStatus {
	successCount := 0
	failureCount := 0
	runningCount := 0

	for _, status := range p.childStatus {
		switch status {
		case Success:
			successCount++
		case Failure:
			failureCount++
		case Running:
			runningCount++
		}
	}

	switch p.policy {
	case RequireOne:
		if successCount > 0 {
			return Success
		}
		if failureCount == len(p.childStatus) {
			return Failure
		}
		return Running

	case RequireAll:
		if failureCount > 0 {
			return Failure
		}
		if successCount == len(p.childStatus) {
			return Success
		}
		return Running

	case SuccessOnAll:
		if successCount == len(p.childStatus) {
			return Success
		}
		return Running

	case SuccessOnOne:
		if successCount > 0 {
			return Success
		}
		return Running

	default:
		// Default to RequireOne
		if successCount > 0 {
			return Success
		}
		if failureCount == len(p.childStatus) {
			return Failure
		}
		return Running
	}
}
