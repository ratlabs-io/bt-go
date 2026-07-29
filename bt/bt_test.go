package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

func alwaysFalseCondition() *bt.Condition {
	return bt.NewCondition(func(env bt.Env) bool {
		return false
	})
}

func alwaysTrueCondition() *bt.Condition {
	return bt.NewCondition(func(env bt.Env) bool {
		return true
	})
}

func alwaysSuccessAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		env.Set("result", "success")
		return bt.Success
	})
}

func alwaysFailureAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		env.Set("result", "failure")
		return bt.Failure
	})
}

func alwaysRunningAction() bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
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
	env := bt.NewEnv(context.Background())
	if alwaysSuccessAction().Tick(env) != bt.Success {
		t.Errorf("expected success")
	}
	if result, ok := env.Get("result"); !ok || result != "success" {
		t.Errorf("expected result to be success")
	}
}

func TestActionNilFunc(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if bt.NewAction(nil).Tick(env) != bt.Failure {
		t.Errorf("nil action func should return Failure")
	}
}

func TestSequence(t *testing.T) {
	env := bt.NewEnv(context.Background())
	sequence := bt.NewSequence(
		alwaysSuccessAction(),
		alwaysFailureAction(),
		alwaysRunningAction(),
	)
	if result := sequence.Tick(env); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysRunningAction(),
	)
	if result := sequence.Tick(env); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	sequence = bt.NewSequence(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysSuccessAction(),
	)
	if result := sequence.Tick(env); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

func TestSequenceReactiveRestart(t *testing.T) {
	// Sequence is reactive: every Tick starts at the first child.
	var firstCalls, secondCalls int
	first := bt.NewAction(func(env bt.Env) bt.RunStatus {
		firstCalls++
		return bt.Success
	})
	second := bt.NewAction(func(env bt.Env) bt.RunStatus {
		secondCalls++
		if secondCalls < 2 {
			return bt.Running
		}
		return bt.Success
	})

	seq := bt.NewSequence(first, second)
	env := bt.NewEnv(context.Background())

	if seq.Tick(env) != bt.Running {
		t.Fatal("expected Running on first tick")
	}
	if seq.Tick(env) != bt.Success {
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
	env := bt.NewEnv(context.Background())
	seq := bt.NewSequence(alwaysSuccessAction(), nil)
	if seq.Tick(env) != bt.Failure {
		t.Errorf("nil child should yield Failure")
	}
}

func TestSequenceEmpty(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if bt.NewSequence().Tick(env) != bt.Success {
		t.Errorf("empty sequence should succeed (vacuous truth)")
	}
}

func TestSelector(t *testing.T) {
	env := bt.NewEnv(context.Background())
	selector := bt.NewSelector(
		alwaysFailureAction(),
		alwaysFailureAction(),
	)
	if result := selector.Tick(env); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	selector = bt.NewSelector(
		alwaysFailureAction(),
		alwaysRunningAction(),
	)
	if result := selector.Tick(env); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}

	selector = bt.NewSelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
	)
	if result := selector.Tick(env); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
}

func TestSelectorPriorityPreempt(t *testing.T) {
	// Higher-priority (earlier) children are re-checked every tick.
	highReady := false
	high := bt.NewAction(func(env bt.Env) bt.RunStatus {
		if highReady {
			return bt.Success
		}
		return bt.Failure
	})
	var lowCalls int
	low := bt.NewAction(func(env bt.Env) bt.RunStatus {
		lowCalls++
		return bt.Running
	})

	sel := bt.NewSelector(high, low)
	env := bt.NewEnv(context.Background())

	if sel.Tick(env) != bt.Running {
		t.Fatal("expected Running while high fails and low runs")
	}
	highReady = true
	if sel.Tick(env) != bt.Success {
		t.Fatal("expected high priority to preempt once ready")
	}
	if lowCalls != 1 {
		t.Errorf("low should not run after preempt, got %d calls", lowCalls)
	}
}

func TestPrioritySelectorAlias(t *testing.T) {
	env := bt.NewEnv(context.Background())
	priority := bt.NewPrioritySelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
	)
	if result := priority.Tick(env); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}
	// Same concrete type as Selector.
	_ = (*bt.Selector)(priority)
}

func TestCondition(t *testing.T) {
	env := bt.NewEnv(context.Background())

	conditional := bt.NewConditional(alwaysTrueCondition(), alwaysSuccessAction())
	if result := conditional.Tick(env); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	conditional = bt.NewConditional(alwaysFalseCondition(), alwaysFailureAction())
	if result := conditional.Tick(env); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
}

func TestConditionLeaf(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if alwaysTrueCondition().Tick(env) != bt.Success {
		t.Error("true condition should Success")
	}
	if alwaysFalseCondition().Tick(env) != bt.Failure {
		t.Error("false condition should Failure")
	}
	if bt.NewCondition(nil).Tick(env) != bt.Failure {
		t.Error("nil check func should Failure")
	}
}

func TestConditionalNilParts(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if bt.NewConditional(nil, alwaysSuccessAction()).Tick(env) != bt.Failure {
		t.Error("nil condition should Failure")
	}
	if bt.NewConditional(alwaysTrueCondition(), nil).Tick(env) != bt.Failure {
		t.Error("nil action with true condition should Failure")
	}
}

func TestBinarySelector(t *testing.T) {
	env := bt.NewEnv(context.Background())
	binary := bt.NewBinarySelector(
		alwaysSuccessAction(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if binary.Tick(env) != bt.Success {
		t.Errorf("expected success")
	}
	binary = bt.NewBinarySelector(
		alwaysFailureAction(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if result := binary.Tick(env); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}
}

func TestBinarySelectorWithCondition(t *testing.T) {
	env := bt.NewEnv(context.Background())
	node := bt.NewBinarySelector(
		alwaysTrueCondition(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if node.Tick(env) != bt.Success {
		t.Error("true condition should take IfTrue branch")
	}
	node = bt.NewBinarySelector(
		alwaysFalseCondition(),
		alwaysSuccessAction(),
		alwaysFailureAction(),
	)
	if node.Tick(env) != bt.Failure {
		t.Error("false condition should take IfFalse branch")
	}
}

func TestSwitch(t *testing.T) {
	switchNode := bt.NewSwitch(
		func(env bt.Env) string {
			key, _ := env.Get("key")
			return key.(string)
		},
		map[string]bt.Behavior{
			"success": alwaysSuccessAction(),
			"failure": alwaysFailureAction(),
		},
		alwaysRunningAction(),
	)

	env := bt.NewEnv(context.Background())

	env.Set("key", "success")
	if result := switchNode.Tick(env); result != bt.Success {
		t.Errorf("expected success, but got %v", result)
	}

	env.Set("key", "failure")
	if result := switchNode.Tick(env); result != bt.Failure {
		t.Errorf("expected failure, but got %v", result)
	}

	env.Set("key", "unknown")
	if result := switchNode.Tick(env); result != bt.Running {
		t.Errorf("expected running, but got %v", result)
	}
}

func TestSwitchNoDefault(t *testing.T) {
	node := bt.NewSwitch(
		func(env bt.Env) string { return "missing" },
		map[string]bt.Behavior{"ok": alwaysSuccessAction()},
		nil,
	)
	env := bt.NewEnv(context.Background())
	if node.Tick(env) != bt.Failure {
		t.Error("missing case with no default should Failure")
	}
}

func TestEnvData(t *testing.T) {
	env := bt.NewEnv(context.Background())
	env.Set("a", 1)
	if !env.Has("a") {
		t.Fatal("Has should be true after Set")
	}
	v, ok := env.Get("a")
	if !ok || v != 1 {
		t.Fatalf("Get = %v, %v", v, ok)
	}
	// Set/Get must use the same store as the blackboard.
	bb := env.Blackboard()
	if bv, ok := bb.Get("a"); !ok || bv != 1 {
		t.Fatal("env data must live on the blackboard")
	}
	env.Delete("a")
	if env.Has("a") {
		t.Fatal("Delete should remove key")
	}
	if bb.Has("a") {
		t.Fatal("Delete via env should clear blackboard")
	}
}

func TestEnvCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	env := bt.NewEnv(parent)
	if env.Context().Err() != nil {
		t.Fatal("context should not be done yet")
	}
	cancel()
	if env.Context().Err() == nil {
		t.Fatal("env.Context should reflect parent cancellation")
	}
}

func TestEnvNilParent(t *testing.T) {
	env := bt.NewEnv(nil)
	if env.Context() == nil {
		t.Fatal("nil parent should fall back to Background")
	}
	env.Set("k", "v")
	if v, ok := env.Get("k"); !ok || v != "v" {
		t.Fatal("env should still store data on blackboard")
	}
}

func TestEnvDoesNotImplementContext(t *testing.T) {
	// Env is has-a context.Context, not is-a. Agent state must not look like
	// request-scoped context values.
	env := bt.NewEnv(context.Background())
	if _, ok := any(env).(context.Context); ok {
		t.Fatal("Env must not implement context.Context")
	}
	// Lifecycle is still available explicitly.
	if env.Context() == nil {
		t.Fatal("Context() must return a usable stdlib context")
	}
}

func TestChildrenProvider(t *testing.T) {
	seq := bt.NewSequence(alwaysSuccessAction(), alwaysFailureAction())
	var cp bt.ChildrenProvider = seq
	if len(cp.GetChildren()) != 2 {
		t.Errorf("expected 2 children")
	}
}
