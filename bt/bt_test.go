package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

func alwaysFalseCondition() *bt.Condition {
	return bt.NewCondition(func(ctx bt.BehaviorContext) bool {
		return false
	})
}

func alwaysTrueCondition() *bt.Condition {
	return bt.NewCondition(func(ctx bt.BehaviorContext) bool {
		return true
	})
}

func alwaysSuccessAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		ctx.Set("result", "success")
		return bt.Success
	})
}

func alwaysFailureAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		ctx.Set("result", "failure")
		return bt.Failure
	})
}

func alwaysRunningAction() bt.Behavior {
	return bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Running
	})
}

func TestRunStatusString(t *testing.T) {
	cases := []struct {
		status bt.RunStatus
		want   string
	}{
		{bt.Success, "Success"},
		{bt.Failure, "Failure"},
		{bt.Running, "Running"},
		{bt.RunStatus(99), "Unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestAction(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if alwaysSuccessAction().Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	if result, ok := ctx.Get("result"); !ok || result != "success" {
		t.Errorf("expected result to be success")
	}
}

func TestActionNilFunc(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if bt.NewAction(nil).Tick(ctx) != bt.Failure {
		t.Errorf("nil action func should return Failure")
	}
}

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

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysRunningAction(),
	)
	if result := sequence.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysSuccessAction(),
	)
	if result := sequence.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

func TestSequenceReactiveRestart(t *testing.T) {
	// Sequence is reactive: every Tick starts at the first child.
	var firstCalls, secondCalls int
	first := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		firstCalls++
		return bt.Success
	})
	second := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		secondCalls++
		if secondCalls < 2 {
			return bt.Running
		}
		return bt.Success
	})

	seq := bt.NewSequence(first, second)
	ctx := bt.NewBehaviorContext(context.Background())

	if seq.Tick(ctx) != bt.Running {
		t.Fatal("expected Running on first tick")
	}
	if seq.Tick(ctx) != bt.Success {
		t.Fatal("expected Success on second tick")
	}
	if firstCalls != 2 {
		t.Errorf("reactive sequence should re-tick first child each time, got %d calls", firstCalls)
	}
	if secondCalls != 2 {
		t.Errorf("expected second child called twice, got %d", secondCalls)
	}
}

func TestSequenceNilChild(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	seq := bt.NewSequence(alwaysSuccessAction(), nil)
	if seq.Tick(ctx) != bt.Failure {
		t.Errorf("nil child should yield Failure")
	}
}

func TestSequenceEmpty(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if bt.NewSequence().Tick(ctx) != bt.Success {
		t.Errorf("empty sequence should succeed (vacuous truth)")
	}
}

func TestSelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	selector := bt.NewSelector(
		alwaysFailureAction(),
		alwaysFailureAction(),
	)
	if result := selector.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
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

func TestSelectorPriorityPreempt(t *testing.T) {
	// Higher-priority (earlier) children are re-checked every tick.
	highReady := false
	high := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		if highReady {
			return bt.Success
		}
		return bt.Failure
	})
	var lowCalls int
	low := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		lowCalls++
		return bt.Running
	})

	sel := bt.NewSelector(high, low)
	ctx := bt.NewBehaviorContext(context.Background())

	if sel.Tick(ctx) != bt.Running {
		t.Fatal("expected Running while high fails and low runs")
	}
	highReady = true
	if sel.Tick(ctx) != bt.Success {
		t.Fatal("expected high priority to preempt once ready")
	}
	if lowCalls != 1 {
		t.Errorf("low should not run after preempt, got %d calls", lowCalls)
	}
}

func TestPrioritySelectorAlias(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	priority := bt.NewPrioritySelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
	)
	if result := priority.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
	// Same concrete type as Selector.
	_ = (*bt.Selector)(priority)
}

func TestCondition(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	conditional := bt.NewConditional(alwaysTrueCondition(), alwaysSuccessAction())
	if result := conditional.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	conditional = bt.NewConditional(alwaysFalseCondition(), alwaysFailureAction())
	if result := conditional.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
}

func TestConditionLeaf(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if alwaysTrueCondition().Tick(ctx) != bt.Success {
		t.Error("true condition should Success")
	}
	if alwaysFalseCondition().Tick(ctx) != bt.Failure {
		t.Error("false condition should Failure")
	}
	if bt.NewCondition(nil).Tick(ctx) != bt.Failure {
		t.Error("nil check func should Failure")
	}
}

func TestConditionalNilParts(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	if bt.NewConditional(nil, alwaysSuccessAction()).Tick(ctx) != bt.Failure {
		t.Error("nil condition should Failure")
	}
	if bt.NewConditional(alwaysTrueCondition(), nil).Tick(ctx) != bt.Failure {
		t.Error("nil action with true condition should Failure")
	}
}

func TestBinarySelector(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	binary := bt.NewBinarySelector(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if binary.Tick(ctx) != bt.Success {
		t.Errorf("expected success")
	}
	binary = bt.NewBinarySelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if result := binary.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
}

func TestBinarySelectorWithCondition(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	node := bt.NewBinarySelector(
		alwaysTrueCondition(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if node.Tick(ctx) != bt.Success {
		t.Error("true condition should take IfTrue branch")
	}
	node = bt.NewBinarySelector(
		alwaysFalseCondition(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if node.Tick(ctx) != bt.Failure {
		t.Error("false condition should take IfFalse branch")
	}
}

func TestSwitch(t *testing.T) {
	switchNode := bt.NewSwitch(
		func(ctx bt.BehaviorContext) string {
			key, _ := ctx.Get("key")
			return key.(string)
		},
		map[string]bt.Behavior{
			"success": alwaysSuccessAction(),
			"failure": alwaysFailureAction(),
		},
		alwaysRunningAction(),
	)

	ctx := bt.NewBehaviorContext(context.Background())

	ctx.Set("key", "success")
	if result := switchNode.Tick(ctx); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	ctx.Set("key", "failure")
	if result := switchNode.Tick(ctx); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	ctx.Set("key", "unknown")
	if result := switchNode.Tick(ctx); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
}

func TestSwitchNoDefault(t *testing.T) {
	node := bt.NewSwitch(
		func(ctx bt.BehaviorContext) string { return "missing" },
		map[string]bt.Behavior{"ok": alwaysSuccessAction()},
		nil,
	)
	ctx := bt.NewBehaviorContext(context.Background())
	if node.Tick(ctx) != bt.Failure {
		t.Error("missing case with no default should Failure")
	}
}

func TestBehaviorContextData(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())
	ctx.Set("a", 1)
	if !ctx.Has("a") {
		t.Fatal("Has should be true after Set")
	}
	v, ok := ctx.Get("a")
	if !ok || v != 1 {
		t.Fatalf("Get = %v, %v", v, ok)
	}
	// Context Set/Get must use the same store as the blackboard.
	bb := ctx.GetBlackboard()
	if bv, ok := bb.Get("a"); !ok || bv != 1 {
		t.Fatal("context data must live on the blackboard")
	}
	ctx.Delete("a")
	if ctx.Has("a") {
		t.Fatal("Delete should remove key")
	}
	if bb.Has("a") {
		t.Fatal("Delete via context should clear blackboard")
	}
}

func TestBehaviorContextCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	ctx := bt.NewBehaviorContext(parent)
	if ctx.Context().Err() != nil {
		t.Fatal("context should not be done yet")
	}
	cancel()
	if ctx.Context().Err() == nil {
		t.Fatal("context should reflect parent cancellation")
	}
}

func TestBehaviorContextNilParent(t *testing.T) {
	ctx := bt.NewBehaviorContext(nil)
	if ctx.Context() == nil {
		t.Fatal("nil parent should fall back to Background")
	}
	ctx.Set("k", "v")
	if v, ok := ctx.Get("k"); !ok || v != "v" {
		t.Fatal("context should still store data")
	}
}

func TestChildrenProvider(t *testing.T) {
	seq := bt.NewSequence(alwaysSuccessAction(), alwaysFailureAction())
	var cp bt.ChildrenProvider = seq
	if len(cp.GetChildren()) != 2 {
		t.Errorf("expected 2 children")
	}
}
