package bt_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ratlabs-io/bt-go/bt"
)

func TestTreeRunnerOptions(t *testing.T) {
	// Create simple success action
	simpleAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})

	// Test default tick rate (100ms)
	runner := bt.NewTreeRunner(simpleAction)
	if runner == nil {
		t.Fatal("TreeRunner should not be nil")
	}

	// Test custom tick rate
	customRate := time.Millisecond * 50
	runner = bt.NewTreeRunner(simpleAction, bt.WithTickRate(customRate))
	if runner == nil {
		t.Fatal("TreeRunner should not be nil")
	}

	// Test with callbacks
	var successCalled, failureCalled, runningCalled bool
	runner = bt.NewTreeRunner(
		simpleAction,
		bt.WithCallbacks(
			func() { successCalled = true },
			func() { failureCalled = true },
			func() { runningCalled = true },
		),
	)

	if runner == nil {
		t.Fatal("TreeRunner should not be nil")
	}

	// Test RunOnce method calls appropriate callback
	bgCtx := context.Background()
	ctx := bt.NewBehaviorContext(bgCtx)
	result := runner.RunOnce(ctx)

	if result != bt.Success {
		t.Errorf("Expected Success, got %v", result)
	}

	if !successCalled {
		t.Error("Success callback should have been called")
	}

	if failureCalled {
		t.Error("Failure callback should not have been called")
	}

	if runningCalled {
		t.Error("Running callback should not have been called")
	}
}

func TestTreeRunnerRunMethod(t *testing.T) {
	// Create a context with cancellation
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := bt.NewBehaviorContext(bgCtx)

	// Create a counter to track the number of ticks
	var counter int32

	// Create an action that increments the counter
	countingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		atomic.AddInt32(&counter, 1)
		return bt.Running
	})

	// Create runner with fast tick rate (1ms)
	runner := bt.NewTreeRunner(countingAction, bt.WithTickRate(time.Millisecond))

	// Start the runner in a goroutine
	go runner.Run(ctx)

	// Wait a bit to allow some ticks to happen
	time.Sleep(50 * time.Millisecond)

	// Cancel the context to stop the runner
	cancel()

	// Wait a bit for the runner to stop
	time.Sleep(10 * time.Millisecond)

	// Check that the counter increased
	if atomic.LoadInt32(&counter) == 0 {
		t.Error("Counter should have been incremented")
	}
}

func TestTreeRunnerWithCallbacks(t *testing.T) {
	// Create a context with cancellation
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := bt.NewBehaviorContext(bgCtx)

	// Create counters for callbacks
	var successCount, failureCount, runningCount int32

	// Create an alternating action
	states := []bt.RunStatus{bt.Success, bt.Failure, bt.Running}
	stateIndex := 0

	alternatingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		result := states[stateIndex]
		stateIndex = (stateIndex + 1) % len(states)
		return result
	})

	// Create runner with callbacks
	runner := bt.NewTreeRunner(
		alternatingAction,
		bt.WithTickRate(time.Millisecond),
		bt.WithCallbacks(
			func() { atomic.AddInt32(&successCount, 1) },
			func() { atomic.AddInt32(&failureCount, 1) },
			func() { atomic.AddInt32(&runningCount, 1) },
		),
	)

	// Start the runner in a goroutine
	go runner.Run(ctx)

	// Wait a bit to allow some ticks (at least one cycle)
	time.Sleep(10 * time.Millisecond)

	// Cancel the context to stop the runner
	cancel()

	// Wait a bit for the runner to stop
	time.Sleep(5 * time.Millisecond)

	// Check that each callback was called at least once
	if atomic.LoadInt32(&successCount) == 0 {
		t.Error("Success callback should have been called")
	}

	if atomic.LoadInt32(&failureCount) == 0 {
		t.Error("Failure callback should have been called")
	}

	if atomic.LoadInt32(&runningCount) == 0 {
		t.Error("Running callback should have been called")
	}
}
