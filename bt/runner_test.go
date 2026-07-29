package bt_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ratlabs-io/bt-go/bt"
)

func TestTreeRunnerOptions(t *testing.T) {
	simpleAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})

	runner := bt.NewTreeRunner(simpleAction)
	if runner == nil {
		t.Fatal("TreeRunner should not be nil")
	}

	customRate := time.Millisecond * 50
	runner = bt.NewTreeRunner(simpleAction, bt.WithTickRate(customRate))
	if runner == nil {
		t.Fatal("TreeRunner should not be nil")
	}

	var successCalled, failureCalled, runningCalled bool
	runner = bt.NewTreeRunner(
		simpleAction,
		bt.WithCallbacks(
			func() { successCalled = true },
			func() { failureCalled = true },
			func() { runningCalled = true },
		),
	)

	ctx := bt.NewBehaviorContext(context.Background())
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

func TestTreeRunnerNilCallbacks(t *testing.T) {
	action := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Failure
	})
	// Nil callbacks must not panic (treated as no-ops / leave defaults).
	runner := bt.NewTreeRunner(action, bt.WithCallbacks(nil, nil, nil))
	ctx := bt.NewBehaviorContext(context.Background())
	if runner.RunOnce(ctx) != bt.Failure {
		t.Fatal("expected Failure")
	}
}

func TestTreeRunnerRunMethod(t *testing.T) {
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := bt.NewBehaviorContext(bgCtx)

	var counter int32
	countingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		atomic.AddInt32(&counter, 1)
		return bt.Running
	})

	runner := bt.NewTreeRunner(countingAction, bt.WithTickRate(time.Millisecond))
	done := make(chan struct{})
	go func() {
		runner.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("runner did not stop after context cancel")
	}

	if atomic.LoadInt32(&counter) == 0 {
		t.Error("Counter should have been incremented")
	}
}

func TestTreeRunnerWithCallbacks(t *testing.T) {
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := bt.NewBehaviorContext(bgCtx)

	var successCount, failureCount, runningCount int32
	states := []bt.RunStatus{bt.Success, bt.Failure, bt.Running}
	var stateIndex int32

	alternatingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		i := atomic.AddInt32(&stateIndex, 1) - 1
		return states[int(i)%len(states)]
	})

	runner := bt.NewTreeRunner(
		alternatingAction,
		bt.WithTickRate(time.Millisecond),
		bt.WithCallbacks(
			func() { atomic.AddInt32(&successCount, 1) },
			func() { atomic.AddInt32(&failureCount, 1) },
			func() { atomic.AddInt32(&runningCount, 1) },
		),
	)

	done := make(chan struct{})
	go func() {
		runner.Run(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("runner did not stop after context cancel")
	}

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

func TestTreeRunnerRunOnceStatuses(t *testing.T) {
	var gotSuccess, gotFailure, gotRunning bool
	makeRunner := func(status bt.RunStatus) *bt.TreeRunner {
		gotSuccess, gotFailure, gotRunning = false, false, false
		return bt.NewTreeRunner(
			bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus { return status }),
			bt.WithCallbacks(
				func() { gotSuccess = true },
				func() { gotFailure = true },
				func() { gotRunning = true },
			),
		)
	}
	ctx := bt.NewBehaviorContext(context.Background())

	if makeRunner(bt.Failure).RunOnce(ctx) != bt.Failure || !gotFailure || gotSuccess || gotRunning {
		t.Error("Failure callback mismatch")
	}
	if makeRunner(bt.Running).RunOnce(ctx) != bt.Running || !gotRunning || gotSuccess || gotFailure {
		t.Error("Running callback mismatch")
	}
}
