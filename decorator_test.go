package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go"
)

func successAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Success
	})
}

func failureAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Failure
	})
}

func runningAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Running
	})
}

func TestInverter(t *testing.T) {
	env := bt.NewEnv(context.Background())

	if result := bt.NewInverter(successAction()).Tick(env); result != bt.Failure {
		t.Errorf("Inverter should convert Success to Failure, got %v", result)
	}
	if result := bt.NewInverter(failureAction()).Tick(env); result != bt.Success {
		t.Errorf("Inverter should convert Failure to Success, got %v", result)
	}
	if result := bt.NewInverter(runningAction()).Tick(env); result != bt.Running {
		t.Errorf("Inverter should not change Running status, got %v", result)
	}
	if result := bt.NewInverter(nil).Tick(env); result != bt.Failure {
		t.Errorf("Inverter with nil child should return Failure, got %v", result)
	}

	inverter := bt.NewInverter(nil)
	inverter.SetChild(successAction())
	if result := inverter.Tick(env); result != bt.Failure {
		t.Errorf("Inverter should convert Success to Failure after SetChild, got %v", result)
	}
	if inverter.GetChild() == nil {
		t.Errorf("GetChild should not return nil after SetChild")
	}

	// Decorator interface.
	var d bt.Decorator = inverter
	if d.GetChild() == nil {
		t.Error("Decorator.GetChild failed")
	}
}

func TestRepeater(t *testing.T) {
	env := bt.NewEnv(context.Background())

	counter := 0
	countingAction := bt.NewAction(func(env bt.Env) bt.RunStatus {
		counter++
		return bt.Success
	})

	repeater := bt.NewRepeater(countingAction, 3)

	if result := repeater.Tick(env); result != bt.Running {
		t.Errorf("Repeater should return Running on first tick, got %v", result)
	}
	if counter != 1 {
		t.Errorf("Action should have been called once, got %d", counter)
	}

	if result := repeater.Tick(env); result != bt.Running {
		t.Errorf("Repeater should return Running on second tick, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Action should have been called twice, got %d", counter)
	}

	if result := repeater.Tick(env); result != bt.Success {
		t.Errorf("Repeater should return Success on third tick, got %v", result)
	}
	if counter != 3 {
		t.Errorf("Action should have been called three times, got %d", counter)
	}

	if result := repeater.Tick(env); result != bt.Running {
		t.Errorf("Repeater should start over and return Running, got %v", result)
	}
	if counter != 4 {
		t.Errorf("Action should have been called four times total, got %d", counter)
	}

	if result := bt.NewRepeater(failureAction(), 3).Tick(env); result != bt.Failure {
		t.Errorf("Repeater should return Failure when child fails, got %v", result)
	}
	if result := bt.NewRepeater(runningAction(), 3).Tick(env); result != bt.Running {
		t.Errorf("Repeater should return Running when child is running, got %v", result)
	}
	if result := bt.NewRepeater(nil, 3).Tick(env); result != bt.Failure {
		t.Errorf("Repeater with nil child should return Failure, got %v", result)
	}

	counter = 0
	infiniteRepeater := bt.NewRepeater(countingAction, 0)
	for i := 0; i < 10; i++ {
		if result := infiniteRepeater.Tick(env); result != bt.Running {
			t.Errorf("Infinite repeater should always return Running, got %v", result)
		}
	}
	if counter != 10 {
		t.Errorf("Action should have been called 10 times in infinite mode, got %d", counter)
	}
}

func TestUntilSuccess(t *testing.T) {
	env := bt.NewEnv(context.Background())

	toggleState := false
	toggleAction := bt.NewAction(func(env bt.Env) bt.RunStatus {
		if toggleState {
			toggleState = false
			return bt.Success
		}
		toggleState = true
		return bt.Failure
	})

	untilSuccess := bt.NewUntilSuccess(toggleAction)
	if result := untilSuccess.Tick(env); result != bt.Running {
		t.Errorf("UntilSuccess should return Running on failure, got %v", result)
	}
	if result := untilSuccess.Tick(env); result != bt.Success {
		t.Errorf("UntilSuccess should return Success when child succeeds, got %v", result)
	}

	if result := bt.NewUntilSuccess(successAction()).Tick(env); result != bt.Success {
		t.Errorf("UntilSuccess should return Success immediately with success action, got %v", result)
	}
	if result := bt.NewUntilSuccess(runningAction()).Tick(env); result != bt.Running {
		t.Errorf("UntilSuccess should return Running when child is running, got %v", result)
	}
	if result := bt.NewUntilSuccess(nil).Tick(env); result != bt.Failure {
		t.Errorf("UntilSuccess with nil child should return Failure, got %v", result)
	}
}

func TestUntilFailure(t *testing.T) {
	env := bt.NewEnv(context.Background())

	toggleState := false
	toggleAction := bt.NewAction(func(env bt.Env) bt.RunStatus {
		if toggleState {
			toggleState = false
			return bt.Failure
		}
		toggleState = true
		return bt.Success
	})

	untilFailure := bt.NewUntilFailure(toggleAction)
	if result := untilFailure.Tick(env); result != bt.Running {
		t.Errorf("UntilFailure should return Running on success, got %v", result)
	}
	if result := untilFailure.Tick(env); result != bt.Success {
		t.Errorf("UntilFailure should return Success when child fails, got %v", result)
	}

	if result := bt.NewUntilFailure(failureAction()).Tick(env); result != bt.Success {
		t.Errorf("UntilFailure should return Success immediately with failure action, got %v", result)
	}
	if result := bt.NewUntilFailure(runningAction()).Tick(env); result != bt.Running {
		t.Errorf("UntilFailure should return Running when child is running, got %v", result)
	}
	if result := bt.NewUntilFailure(nil).Tick(env); result != bt.Failure {
		t.Errorf("UntilFailure with nil child should return Failure, got %v", result)
	}
}
