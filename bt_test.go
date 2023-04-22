package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/go-bt"
)

// TestAction tests the Action behavior.
func TestAction(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	action := bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			ctx.Set("result", "success")
			return bt.Success
		},
	}
	if action.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	if result, ok := ctx.Get("result"); !ok || result != "success" {
		t.Errorf("expected result to be success")
	}
}

// TestSequence tests the Sequence behavior.
func TestSequence(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Success
		},
	}
	failure := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Failure
		},
	}
	running := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Running
		},
	}
	sequence := &bt.SequenceNode{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				success,
				failure,
				running,
			},
		},
	}
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
	if result := sequence.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	sequence = &bt.SequenceNode{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				success,
				success,
				running,
			},
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

	sequence = &bt.SequenceNode{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				success,
				success,
				success,
			},
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
	success := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Success
		},
	}
	failure := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Failure
		},
	}
	running := &bt.Action{
		Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
			return bt.Running
		},
	}
	selector := &bt.Selector{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				failure,
				failure,
			},
		},
	}
	if result := selector.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failue, but got %v", result)
	}

	selector = &bt.Selector{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				failure,
				running,
			},
		},
	}
	if result := selector.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	selector = &bt.Selector{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				failure,
				success,
			},
		},
	}
	if result := selector.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

// TestPrioritySelector tests the PrioritySelector behavior.
func TestPrioritySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return true
			},
		},
	}
	failure := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return false
			},
		},
	}
	priority := &bt.PrioritySelector{
		Composite: bt.Composite{
			Children: []bt.Behavior{
				failure,
				success,
			},
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
	if condition.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
}

// TestBinarySelector tests the BinarySelector behavior.
func TestBinarySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return true
			},
		},
	}
	failure := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return false
			},
		},
	}
	binary := &bt.BinarySelector{
		Condition: success,
		IfTrue:    success,
		IfFalse:   failure,
	}
	if binary.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	binary = &bt.BinarySelector{
		Condition: failure,
		IfTrue:    success,
		IfFalse:   failure,
	}
	if binary.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
}

// TestSwitch tests the Switch behavior.
func TestSwitch(t *testing.T) {
	success := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return true
			},
		},
	}
	failure := &bt.ConditionBehavior{
		Condition: &bt.Condition{
			Check: func(ctx *bt.BehaviorContext) bool {
				return false
			},
		},
	}

	switchNode := &bt.Switch{
		Key: "success",
		Cases: map[bt.Key]bt.Behavior{
			"success": success,
			"failure": failure,
		},
		Default: &bt.Action{
			Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
				return bt.Running
			},
		},
	}

	ctx := bt.NewBehaviorContext(context.Background())

	if switchNode.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}

	switchNode.Key = "failure"

	if switchNode.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}

	switchNode.Key = "unknown"

	if switchNode.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
}
