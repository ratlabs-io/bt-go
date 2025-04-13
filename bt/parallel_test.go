package bt_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ratlabs-io/bt-go/bt"
)

// Creates an action that increments a counter and returns a specified status
func createCountingAction(counter *int, mutex *sync.Mutex, delay time.Duration, status bt.RunStatus) bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		// Add a small delay to simulate work
		if delay > 0 {
			time.Sleep(delay)
		}

		mutex.Lock()
		*counter++
		mutex.Unlock()

		return status
	})
}

func TestParallelRequireOne(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Test with all failures - should return Failure
	parallel := bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)

	if result := parallel.Tick(ctx); result != bt.Failure {
		t.Errorf("Expected Failure when all children fail, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one success - should return Success
	parallel = bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)

	if result := parallel.Tick(ctx); result != bt.Success {
		t.Errorf("Expected Success when at least one child succeeds, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one running - should return Running
	parallel = bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when at least one child is running, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelRequireAll(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Test with one failure - should return Failure
	parallel := bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)

	if result := parallel.Tick(ctx); result != bt.Failure {
		t.Errorf("Expected Failure when any child fails, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with all success - should return Success
	parallel = bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)

	if result := parallel.Tick(ctx); result != bt.Success {
		t.Errorf("Expected Success when all children succeed, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one running - should return Running
	parallel = bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when any child is running, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelSuccessOnAll(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Test with one failure - should return Running
	parallel := bt.NewParallel(bt.SuccessOnAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when any child fails, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with all success - should return Success
	parallel = bt.NewParallel(bt.SuccessOnAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)

	if result := parallel.Tick(ctx); result != bt.Success {
		t.Errorf("Expected Success when all children succeed, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one running - should return Running
	parallel = bt.NewParallel(bt.SuccessOnAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when any child is running, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelSuccessOnOne(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Test with all failures - should return Running
	parallel := bt.NewParallel(bt.SuccessOnOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when all children fail, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one success - should return Success
	parallel = bt.NewParallel(bt.SuccessOnOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)

	if result := parallel.Tick(ctx); result != bt.Success {
		t.Errorf("Expected Success when at least one child succeeds, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	// Reset counter
	counter = 0

	// Test with one running - should return Running
	parallel = bt.NewParallel(bt.SuccessOnOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)

	if result := parallel.Tick(ctx); result != bt.Running {
		t.Errorf("Expected Running when no child succeeds and any is running, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelNilChild(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Test with one nil child
	parallel := bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		nil,
	)

	if result := parallel.Tick(ctx); result != bt.Failure {
		t.Errorf("Expected Failure when a child is nil with RequireAll policy, got %v", result)
	}
	if counter != 1 {
		t.Errorf("Expected only valid actions to run, got counter = %d", counter)
	}
}

func TestParallelConcurrency(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var counter int
	var mu sync.Mutex

	// Create actions with different delays
	parallel := bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 50*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 30*time.Millisecond, bt.Success),
	)

	start := time.Now()
	result := parallel.Tick(ctx)
	elapsed := time.Since(start)

	if result != bt.Success {
		t.Errorf("Expected Success, got %v", result)
	}

	// If truly parallel, this should take just over 50ms (the longest task)
	// not 90ms (the sum of all tasks)
	if elapsed > 90*time.Millisecond {
		t.Errorf("Execution took too long (%v), suggesting sequential execution", elapsed)
	}

	if counter != 3 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelStateTracking(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	var mu sync.Mutex

	// Create a stateful action that changes from Running to Success
	stateValue := 0
	statefulAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		mu.Lock()
		defer mu.Unlock()

		if stateValue == 0 {
			stateValue = 1
			return bt.Running
		}
		return bt.Success
	})

	// Counter for the second action - separate from the statefulAction state
	var actionCount int32
	countingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		atomic.AddInt32(&actionCount, 1)
		return bt.Running
	})

	// Create a parallel node with the stateful action and a success action
	parallel := bt.NewParallel(bt.RequireAll,
		statefulAction,
		countingAction,
	)

	// First tick - should return Running because both actions return Running
	result := parallel.Tick(ctx)
	if result != bt.Running {
		t.Errorf("Expected Running on first tick, got %v", result)
	}

	// Check state after first tick
	if stateValue != 1 {
		t.Errorf("Expected stateValue to be 1 after first tick, got %d", stateValue)
	}

	// Second tick - statefulAction should change to Success, countingAction still Running
	result = parallel.Tick(ctx)
	if result != bt.Running {
		t.Errorf("Expected Running on second tick, got %v", result)
	}

	// Verify both actions were called on each tick
	if count := atomic.LoadInt32(&actionCount); count != 2 {
		t.Errorf("Expected countingAction to be called twice, got %d", count)
	}
}
