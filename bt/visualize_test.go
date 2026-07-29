package bt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

type CustomVisualizerAction struct {
	name string
}

func NewCustomVisualizerAction(name string) *CustomVisualizerAction {
	return &CustomVisualizerAction{name: name}
}

func (c *CustomVisualizerAction) Tick(ctx bt.Env) bt.RunStatus {
	return bt.Success
}

func (c *CustomVisualizerAction) VisualizeNode() string {
	return "CustomAction: " + c.name
}

func TestTreeVisualizerBasic(t *testing.T) {
	action1 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	action2 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure })
	sequence := bt.NewSequence(action1, action2)

	result := bt.NewTreeVisualizer(sequence).Visualize()

	if !strings.Contains(result, "Sequence") {
		t.Errorf("Visualization should contain 'Sequence', got: %s", result)
	}
	if !strings.Contains(result, "Action") {
		t.Errorf("Visualization should contain 'Action', got: %s", result)
	}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 3 {
		t.Errorf("Visualization should have at least 3 lines, got %d", len(lines))
	}
}

func TestTreeVisualizerComplex(t *testing.T) {
	action1 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	condition := bt.NewCondition(func(env bt.Env) bool { return true })
	inverter := bt.NewInverter(condition)
	repeater := bt.NewRepeater(action1, 3)
	selector := bt.NewSelector(inverter, repeater)

	result := bt.NewTreeVisualizer(selector).Visualize()

	for _, want := range []string{"Selector", "Inverter", "Repeater", "Condition", "Action", "Repeat count"} {
		if !strings.Contains(result, want) {
			t.Errorf("Visualization should contain %q, got: %s", want, result)
		}
	}
}

func TestTreeVisualizerBinaryAndSwitch(t *testing.T) {
	bin := bt.NewBinarySelector(
		bt.NewCondition(func(env bt.Env) bool { return true }),
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure }),
	)
	sw := bt.NewSwitch(
		func(env bt.Env) string { return "a" },
		map[string]bt.Behavior{
			"a": bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		},
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure }),
	)
	root := bt.NewSequence(bin, sw)
	result := bt.NewTreeVisualizer(root).Visualize()

	for _, want := range []string{"BinarySelector", "Switch", "case", "default", "Condition", "Action"} {
		if !strings.Contains(result, want) {
			t.Errorf("Visualization should contain %q, got: %s", want, result)
		}
	}
}

func TestTreeVisualizerParallel(t *testing.T) {
	p := bt.NewParallel(bt.RequireAll,
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
	)
	result := bt.NewTreeVisualizer(p).Visualize()
	if !strings.Contains(result, "Parallel(RequireAll)") {
		t.Errorf("expected Parallel policy in label, got: %s", result)
	}
}

func TestTreeVisualizerWithStatus(t *testing.T) {
	action1 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	action2 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure })
	sequence := bt.NewSequence(action1, action2)

	statusMap := map[bt.Behavior]bt.RunStatus{
		sequence: bt.Failure,
		action1:  bt.Success,
		action2:  bt.Failure,
	}

	result := bt.NewTreeVisualizer(sequence).WithNodeStatuses(statusMap).Visualize()
	if !strings.Contains(result, "[Success]") {
		t.Errorf("Visualization should contain '[Success]', got: %s", result)
	}
	if !strings.Contains(result, "[Failure]") {
		t.Errorf("Visualization should contain '[Failure]', got: %s", result)
	}
}

func TestCustomNodeVisualizer(t *testing.T) {
	customAction := NewCustomVisualizerAction("TestAction")
	result := bt.NewTreeVisualizer(customAction).Visualize()
	if !strings.Contains(result, "CustomAction: TestAction") {
		t.Errorf("Visualization should contain custom text, got: %s", result)
	}
}

func TestStatusRecorder(t *testing.T) {
	env := bt.NewEnv(context.Background())

	action1 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	action2 := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure })
	sequence := bt.NewSequence(action1, action2)

	snapshot := bt.NewStatusRecorder()
	status1 := snapshot.Tick(env, action1)
	status2 := snapshot.Tick(env, action2)

	if status1 != bt.Success {
		t.Errorf("Expected Success status from action1, got %v", status1)
	}
	if status2 != bt.Failure {
		t.Errorf("Expected Failure status from action2, got %v", status2)
	}

	statusMap := snapshot.GetStatusMap()
	if statusMap[action1] != bt.Success {
		t.Errorf("Status map should contain Success for action1, got %v", statusMap[action1])
	}
	if statusMap[action2] != bt.Failure {
		t.Errorf("Status map should contain Failure for action2, got %v", statusMap[action2])
	}

	result := snapshot.Visualize(sequence)
	if strings.Contains(result, "Sequence [") {
		t.Errorf("Visualization should not contain status for sequence, got: %s", result)
	}
}

func TestSaveTreeSnapshotDeprecatedAlias(t *testing.T) {
	// Deprecated constructor still works.
	rec := bt.NewSaveTreeSnapshot()
	if rec == nil {
		t.Fatal("NewSaveTreeSnapshot should return a recorder")
	}
}
