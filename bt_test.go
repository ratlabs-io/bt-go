package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/go-bt"
)

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
	sequence := &bt.Sequence{
		Children: []bt.Behavior{
			alwaysSuccessAction(),
			alwaysFailureAction(),
			alwaysRunningAction(),
		},
	}
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	sequence = &bt.Sequence{
		Children: []bt.Behavior{
			alwaysSuccessAction(),
			alwaysSuccessAction(),
			alwaysRunningAction(),
		},
	}

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

	sequence = &bt.Sequence{
		Children: []bt.Behavior{
			alwaysSuccessAction(),
			alwaysSuccessAction(),
			alwaysSuccessAction(),
		},
	}

	ctx.Set("running", false)
	if result := sequence.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestSelector tests the Selector behavior.
func TestSelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	selector := &bt.Selector{
		Children: []bt.Behavior{
			alwaysFailureAction(),
			alwaysFailureAction(),
		},
	}
	if result := selector.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failue, but got %v", result)
	}

	selector = &bt.Selector{
		Children: []bt.Behavior{
			alwaysFailureAction(),
			alwaysRunningAction(),
		},
	}
	if result := selector.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	selector = &bt.Selector{
		Children: []bt.Behavior{
			alwaysFailureAction(),
			alwaysSuccessAction(),
		},
	}
	if result := selector.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestPrioritySelector tests the PrioritySelector behavior.
func TestPrioritySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	priority := &bt.PrioritySelector{
		Children: []bt.Behavior{
			alwaysFailureAction(),
			alwaysSuccessAction(),
		},
	}
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
	condition := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return true
			},
		},
	}
	if condition.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	condition = &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return false
			},
		},
	}
	if result := condition.Tick(ctx); result != bt.Failure {
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
		Key: "success",
		Cases: map[bt.Key]bt.Behavior{
			"success": alwaysSuccessAction(),
			"failure": alwaysFailureAction(),
		},
		Default: alwaysRunningAction(),
	}

	ctx := bt.NewBehaviorContext(context.Background())

	if result := switchNode.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	switchNode.Key = "failure"

	if result := switchNode.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	switchNode.Key = "unknown"

	if result := switchNode.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
}
