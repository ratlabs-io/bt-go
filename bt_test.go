package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/go-bt"
)

func alwaysFalseCondition() *bt.Condition {
	return &bt.Condition{
		CheckFunc: func(ctx *bt.BehaviorContext) bool {
			return false
		},
	}
}

func alwaysTrueCondition() *bt.Condition {
	return &bt.Condition{
		CheckFunc: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
}

func alwaysSuccessAction() bt.Behavior {
	return bt.NewAction(func(ctx *bt.BehaviorContext) bt.RunStatus {
		ctx.Set("result", "success")
		return bt.Success
	})
}

func alwaysFailureAction() bt.Behavior {
	return bt.NewAction(func(ctx *bt.BehaviorContext) bt.RunStatus {
		ctx.Set("result", "failure")
		return bt.Failure
	})
}

func alwaysRunningAction() bt.Behavior {
	return bt.NewAction(func(ctx *bt.BehaviorContext) bt.RunStatus {
		return bt.Running
	})
}

// TestAction tests the Action behavior.
func TestAction(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if alwaysSuccessAction().Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	if result, ok := ctx.Get("result"); !ok || result != "success" {
		t.Errorf("expected result to be success")
	}
}

// TestSequence tests the Sequence behavior.
func TestSequence(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	sequence := bt.NewSequence(
		alwaysSuccessAction(),
		alwaysFailureAction(),
		alwaysRunningAction(),
	)
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysRunningAction(),
	)

	if result := sequence.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
	if result := sequence.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
	ctx.Set("running", true)
	if result := sequence.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysSuccessAction(),
	)

	ctx.Set("running", false)
	if result := sequence.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestSelector tests the Selector behavior.
func TestSelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	selector := bt.NewSelector(
		alwaysFailureAction(),
		alwaysFailureAction(),
	)

	if result := selector.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failue, but got %v", result)
	}

	selector = bt.NewSelector(
		alwaysFailureAction(),
		alwaysRunningAction(),
	)
	if result := selector.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	selector = bt.NewSelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
	)
	if result := selector.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestPrioritySelector tests the PrioritySelector behavior.
func TestPrioritySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	priority := bt.NewPrioritySelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
	)
	if result := priority.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
	if result := priority.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestCondition tests the Condition behavior.
func TestCondition(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Case 1: condition evaluates to true
	conditional := bt.NewConditional(
		alwaysTrueCondition(),
		alwaysSuccessAction())

	if result := conditional.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	// Case 2: condition evaluates to false
	conditional = bt.NewConditional(
		alwaysFalseCondition(),
		alwaysFailureAction())

	if result := conditional.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

}

// TestBinarySelector tests the BinarySelector behavior.
func TestBinarySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	binary := &bt.BinarySelector{
		Condition: alwaysSuccessAction(),
		IfTrue:    alwaysSuccessAction(),
		IfFalse:   alwaysFailureAction(),
	}
	if binary.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	binary = &bt.BinarySelector{
		Condition: alwaysFailureAction(),
		IfTrue:    alwaysSuccessAction(),
		IfFalse:   alwaysFailureAction(),
	}
	if result := binary.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
}

// TestSwitch tests the Switch behavior.
func TestSwitch(t *testing.T) {
	switchNode := &bt.Switch{
		KeyFunc: func(ctx *bt.BehaviorContext) string {
			key, _ := ctx.Get("key")
			return key.(string)
		},
		Cases: map[string]bt.Behavior{
			"success": alwaysSuccessAction(),
			"failure": alwaysFailureAction(),
		},
		Default: alwaysRunningAction(),
	}

	ctx := bt.NewBehaviorContext(context.Background())

	// Test with success key
	ctx.Set("key", "success")
	if result := switchNode.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	// Test with failure key
	ctx.Set("key", "failure")
	if result := switchNode.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	// Test with unknown key
	ctx.Set("key", "unknown")
	if result := switchNode.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
}
