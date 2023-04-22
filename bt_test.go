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
	success := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
	failure := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return false
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
	if sequence.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
	if sequence.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
	if sequence.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	if sequence.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	ctx.Set("running", true)
	if sequence.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	ctx.Set("running", false)
	if sequence.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
}

// TestSelector tests the Selector behavior.
func TestSelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
	failure := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return false
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
				running,
				success,
			},
		},
	}
	if selector.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	if selector.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	if selector.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
}

// TestPrioritySelector tests the PrioritySelector behavior.
func TestPrioritySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
	failure := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return false
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
	if priority.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	if priority.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
}

// TestCondition tests the Condition behavior.
func TestCondition(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	condition := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
	if condition.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	condition = &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return false
		},
	}
	if condition.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
}

// TestBinarySelector tests the BinarySelector behavior.
func TestBinarySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	success := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return true
		},
	}
	failure := &bt.Condition{
		Condition: func(ctx *bt.BehaviorContext) bool {
			return false
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
	ctx := bt.NewBehaviorContext(context.Background())
	switchNode := &bt.Switch{
		Values: map[interface{}]bt.Behavior{
			"success": &bt.Condition{
				Condition: func(ctx *bt.BehaviorContext) bool {
					return true
				},
			},
			"failure": &bt.Condition{
				Condition: func(ctx *bt.BehaviorContext) bool {
					return false
				},
			},
		},
		Default: &bt.Action{
			Action: func(ctx *bt.BehaviorContext) bt.RunStatus {
				return bt.Running
			},
		},
		ValueFunc: func(ctx *bt.BehaviorContext) interface{} {
			return ctx.Get("value")
		},
	}
	if switchNode.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
	ctx.Set("value", "success")
	if switchNode.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	ctx.Set("value", "failure")
	if switchNode.Tick(ctx) != bt.Failure {
		t.Errorf("expected failure")
	}
	ctx.Set("value", "unknown")
	if switchNode.Tick(ctx) != bt.Running {
		t.Errorf("expected running")
	}
}
