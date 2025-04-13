package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

// Helper functions for testing
func successAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})
}

func failureAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Failure
	})
}

func runningAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Running
	})
}

func TestInverter(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Test inverting Success to Failure
	inverter := bt.NewInverter(successAction())
	if result := inverter.Tick(ctx); result != bt.Failure {
		t.Errorf("Inverter should convert Success to Failure, got %v", result)
	}

	// Test inverting Failure to Success
	inverter = bt.NewInverter(failureAction())
	if result := inverter.Tick(ctx); result != bt.Success {
		t.Errorf("Inverter should convert Failure to Success, got %v", result)
	}

	// Test not changing Running status
	inverter = bt.NewInverter(runningAction())
	if result := inverter.Tick(ctx); result != bt.Running {
		t.Errorf("Inverter should not change Running status, got %v", result)
	}

	// Test with nil child
	inverter = bt.NewInverter(nil)
	if result := inverter.Tick(ctx); result != bt.Failure {
		t.Errorf("Inverter with nil child should return Failure, got %v", result)
	}

	// Test SetChild and GetChild
	inverter = bt.NewInverter(nil)
	inverter.SetChild(successAction())
	if result := inverter.Tick(ctx); result != bt.Failure {
		t.Errorf("Inverter should convert Success to Failure after SetChild, got %v", result)
	}

	if inverter.GetChild() == nil {
		t.Errorf("GetChild should not return nil after SetChild")
	}
}

func TestRepeater(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Test repeating a specific number of times
	counter := 0
	countingAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		counter++
		return bt.Success
	})

	repeater := bt.NewRepeater(countingAction, 3)

	// First tick - should run once and return Running
	if result := repeater.Tick(ctx); result != bt.Running {
		t.Errorf("Repeater should return Running on first tick, got %v", result)
	}
	if counter != 1 {
		t.Errorf("Action should have been called once, got %d", counter)
	}

	// Second tick - should run again and return Running
	if result := repeater.Tick(ctx); result != bt.Running {
		t.Errorf("Repeater should return Running on second tick, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Action should have been called twice, got %d", counter)
	}

	// Third tick - should run one last time and return Success
	if result := repeater.Tick(ctx); result != bt.Success {
		t.Errorf("Repeater should return Success on third tick, got %v", result)
	}
	if counter != 3 {
		t.Errorf("Action should have been called three times, got %d", counter)
	}

	// Fourth tick - should start over
	if result := repeater.Tick(ctx); result != bt.Running {
		t.Errorf("Repeater should start over and return Running, got %v", result)
	}
	if counter != 4 {
		t.Errorf("Action should have been called four times total, got %d", counter)
	}

	// Test with failure action
	repeater = bt.NewRepeater(failureAction(), 3)
	if result := repeater.Tick(ctx); result != bt.Failure {
		t.Errorf("Repeater should return Failure when child fails, got %v", result)
	}

	// Test with running action
	repeater = bt.NewRepeater(runningAction(), 3)
	if result := repeater.Tick(ctx); result != bt.Running {
		t.Errorf("Repeater should return Running when child is running, got %v", result)
	}

	// Test with nil child
	repeater = bt.NewRepeater(nil, 3)
	if result := repeater.Tick(ctx); result != bt.Failure {
		t.Errorf("Repeater with nil child should return Failure, got %v", result)
	}

	// Test infinite repeater (count <= 0)
	counter = 0
	infiniteRepeater := bt.NewRepeater(countingAction, 0)
	for i := 0; i < 10; i++ {
		if result := infiniteRepeater.Tick(ctx); result != bt.Running {
			t.Errorf("Infinite repeater should always return Running, got %v", result)
		}
	}
	if counter != 10 {
		t.Errorf("Action should have been called 10 times in infinite mode, got %d", counter)
	}
}

func TestUntilSuccess(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Test with alternating success/failure
	toggleState := false
	toggleAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		if toggleState {
			toggleState = false
			return bt.Success
		}
		toggleState = true
		return bt.Failure
	})

	untilSuccess := bt.NewUntilSuccess(toggleAction)

	// First tick - should get failure and return Running
	if result := untilSuccess.Tick(ctx); result != bt.Running {
		t.Errorf("UntilSuccess should return Running on failure, got %v", result)
	}

	// Second tick - should get success and return Success
	if result := untilSuccess.Tick(ctx); result != bt.Success {
		t.Errorf("UntilSuccess should return Success when child succeeds, got %v", result)
	}

	// Test with success action
	untilSuccess = bt.NewUntilSuccess(successAction())
	if result := untilSuccess.Tick(ctx); result != bt.Success {
		t.Errorf("UntilSuccess should return Success immediately with success action, got %v", result)
	}

	// Test with running action
	untilSuccess = bt.NewUntilSuccess(runningAction())
	if result := untilSuccess.Tick(ctx); result != bt.Running {
		t.Errorf("UntilSuccess should return Running when child is running, got %v", result)
	}

	// Test with nil child
	untilSuccess = bt.NewUntilSuccess(nil)
	if result := untilSuccess.Tick(ctx); result != bt.Failure {
		t.Errorf("UntilSuccess with nil child should return Failure, got %v", result)
	}
}

func TestUntilFailure(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Test with alternating success/failure
	toggleState := false
	toggleAction := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		if toggleState {
			toggleState = false
			return bt.Failure
		}
		toggleState = true
		return bt.Success
	})

	untilFailure := bt.NewUntilFailure(toggleAction)

	// First tick - should get success and return Running
	if result := untilFailure.Tick(ctx); result != bt.Running {
		t.Errorf("UntilFailure should return Running on success, got %v", result)
	}

	// Second tick - should get failure and return Success
	if result := untilFailure.Tick(ctx); result != bt.Success {
		t.Errorf("UntilFailure should return Success when child fails, got %v", result)
	}

	// Test with failure action
	untilFailure = bt.NewUntilFailure(failureAction())
	if result := untilFailure.Tick(ctx); result != bt.Success {
		t.Errorf("UntilFailure should return Success immediately with failure action, got %v", result)
	}

	// Test with running action
	untilFailure = bt.NewUntilFailure(runningAction())
	if result := untilFailure.Tick(ctx); result != bt.Running {
		t.Errorf("UntilFailure should return Running when child is running, got %v", result)
	}

	// Test with nil child
	untilFailure = bt.NewUntilFailure(nil)
	if result := untilFailure.Tick(ctx); result != bt.Failure {
		t.Errorf("UntilFailure with nil child should return Failure, got %v", result)
	}
}
