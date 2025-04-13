package bt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

// Custom action that implements NodeVisualizer
type CustomVisualizerAction struct {
	name string
}

func NewCustomVisualizerAction(name string) *CustomVisualizerAction {
	return &CustomVisualizerAction{name: name}
}

func (c *CustomVisualizerAction) Tick(ctx bt.BehaviorContext) bt.RunStatus {
	return bt.Success
}

func (c *CustomVisualizerAction) VisualizeNode() string {
	return "CustomAction: " + c.name
}

func TestTreeVisualizerBasic(t *testing.T) {
	// Create a simple tree
	action1 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})
	action2 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Failure
	})

	sequence := bt.NewSequence(action1, action2)

	// Create visualizer
	visualizer := bt.NewTreeVisualizer(sequence)

	// Get visualization
	result := visualizer.Visualize()

	// Check that the output contains expected node types
	if !strings.Contains(result, "Sequence") {
		t.Errorf("Visualization should contain 'Sequence', got: %s", result)
	}

	if !strings.Contains(result, "Action") {
		t.Errorf("Visualization should contain 'Action', got: %s", result)
	}

	// Check that the output has the expected tree structure
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("Visualization should have at least 3 lines, got %d", len(lines))
	}
}

func TestTreeVisualizerComplex(t *testing.T) {
	// Create a more complex tree
	action1 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})
	condition := bt.NewCondition(func(ctx bt.BehaviorContext) bool {
		return true
	})

	inverter := bt.NewInverter(condition)
	repeater := bt.NewRepeater(action1, 3)

	selector := bt.NewSelector(inverter, repeater)

	// Create visualizer
	visualizer := bt.NewTreeVisualizer(selector)

	// Get visualization
	result := visualizer.Visualize()

	// Check that the output contains expected node types
	if !strings.Contains(result, "Selector") {
		t.Errorf("Visualization should contain 'Selector', got: %s", result)
	}

	if !strings.Contains(result, "Inverter") {
		t.Errorf("Visualization should contain 'Inverter', got: %s", result)
	}

	if !strings.Contains(result, "Repeater") {
		t.Errorf("Visualization should contain 'Repeater', got: %s", result)
	}

	if !strings.Contains(result, "Condition") {
		t.Errorf("Visualization should contain 'Condition', got: %s", result)
	}

	if !strings.Contains(result, "Action") {
		t.Errorf("Visualization should contain 'Action', got: %s", result)
	}

	// Check that Repeater shows count information
	if !strings.Contains(result, "Repeat count") {
		t.Errorf("Visualization should contain 'Repeat count' for Repeater, got: %s", result)
	}
}

func TestTreeVisualizerWithStatus(t *testing.T) {
	// Create a simple tree
	action1 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})
	action2 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Failure
	})

	sequence := bt.NewSequence(action1, action2)

	// Create status map
	statusMap := map[bt.Behavior]bt.RunStatus{
		sequence: bt.Failure,
		action1:  bt.Success,
		action2:  bt.Failure,
	}

	// Create visualizer with status
	visualizer := bt.NewTreeVisualizer(sequence).WithNodeStatuses(statusMap)

	// Get visualization
	result := visualizer.Visualize()

	// Check that status is shown
	if !strings.Contains(result, "[Success]") {
		t.Errorf("Visualization should contain '[Success]', got: %s", result)
	}

	if !strings.Contains(result, "[Failure]") {
		t.Errorf("Visualization should contain '[Failure]', got: %s", result)
	}
}

func TestCustomNodeVisualizer(t *testing.T) {
	// Create a custom visualizer action
	customAction := NewCustomVisualizerAction("TestAction")

	// Create visualizer
	visualizer := bt.NewTreeVisualizer(customAction)

	// Get visualization
	result := visualizer.Visualize()

	// Check that custom text is shown
	if !strings.Contains(result, "CustomAction: TestAction") {
		t.Errorf("Visualization should contain custom text, got: %s", result)
	}
}

func TestSaveTreeSnapshot(t *testing.T) {
	ctx := bt.NewBehaviorContext(context.Background())

	// Create a tree
	action1 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Success
	})
	action2 := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
		return bt.Failure
	})

	sequence := bt.NewSequence(action1, action2)

	// Create snapshot
	snapshot := bt.NewSaveTreeSnapshot()

	// Run snapshot tick on both actions
	status1 := snapshot.Tick(ctx, action1)
	status2 := snapshot.Tick(ctx, action2)

	// Check correct statuses are returned
	if status1 != bt.Success {
		t.Errorf("Expected Success status from action1, got %v", status1)
	}

	if status2 != bt.Failure {
		t.Errorf("Expected Failure status from action2, got %v", status2)
	}

	// Get status map
	statusMap := snapshot.GetStatusMap()

	// Check map contains both actions with correct statuses
	if statusMap[action1] != bt.Success {
		t.Errorf("Status map should contain Success for action1, got %v", statusMap[action1])
	}

	if statusMap[action2] != bt.Failure {
		t.Errorf("Status map should contain Failure for action2, got %v", statusMap[action2])
	}

	// Check visualization
	result := snapshot.Visualize(sequence)

	// Visualization should not show sequence status (not ticked)
	if strings.Contains(result, "Sequence [") {
		t.Errorf("Visualization should not contain status for sequence, got: %s", result)
	}
}
