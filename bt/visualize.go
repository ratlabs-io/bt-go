package bt

import (
	"fmt"
	"reflect"
	"strings"
)

// NodeVisualizer is an interface for custom node visualization.
type NodeVisualizer interface {
	// VisualizeNode returns a string representation of the node.
	VisualizeNode() string
}

// TreeVisualizer generates a text representation of a behavior tree.
type TreeVisualizer struct {
	root           Behavior
	showNodeStatus bool
	nodeStatuses   map[Behavior]RunStatus
}

// NewTreeVisualizer creates a new TreeVisualizer for the given root behavior.
func NewTreeVisualizer(root Behavior) *TreeVisualizer {
	return &TreeVisualizer{
		root:           root,
		showNodeStatus: false,
		nodeStatuses:   make(map[Behavior]RunStatus),
	}
}

// WithNodeStatuses enables status display and sets the node statuses to display.
func (tv *TreeVisualizer) WithNodeStatuses(statuses map[Behavior]RunStatus) *TreeVisualizer {
	tv.showNodeStatus = true
	tv.nodeStatuses = statuses
	return tv
}

// Visualize returns a string representation of the behavior tree.
func (tv *TreeVisualizer) Visualize() string {
	var builder strings.Builder
	tv.visualizeNode(&builder, tv.root, "", true, true)
	return builder.String()
}

// visualizeNode recursively builds a string representation of a behavior tree node.
func (tv *TreeVisualizer) visualizeNode(builder *strings.Builder, node Behavior, prefix string, isLast bool, isRoot bool) {
	if node == nil {
		return
	}

	// Determine the connector and new prefix for child nodes
	connector := "├── "
	newPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		newPrefix = prefix + "    "
	}
	if isRoot {
		connector = ""
	} else {
		builder.WriteString(prefix)
		builder.WriteString(connector)
	}

	// Get node name
	nodeName := tv.getNodeName(node)

	// Add status if enabled
	if tv.showNodeStatus {
		if status, ok := tv.nodeStatuses[node]; ok {
			nodeName = fmt.Sprintf("%s [%s]", nodeName, status)
		}
	}

	builder.WriteString(nodeName)
	builder.WriteString("\n")

	// Handle child nodes
	switch n := node.(type) {
	case *Sequence:
		for i, child := range n.Children {
			tv.visualizeNode(builder, child, newPrefix, i == len(n.Children)-1, false)
		}
	case *Selector:
		for i, child := range n.Children {
			tv.visualizeNode(builder, child, newPrefix, i == len(n.Children)-1, false)
		}
	case *PrioritySelector:
		for i, child := range n.Children {
			tv.visualizeNode(builder, child, newPrefix, i == len(n.Children)-1, false)
		}
	case *Parallel:
		for i, child := range n.Children {
			tv.visualizeNode(builder, child, newPrefix, i == len(n.Children)-1, false)
		}
	case *Conditional:
		tv.visualizeNode(builder, n.Condition, newPrefix, false, false)
		tv.visualizeNode(builder, n.Action, newPrefix, true, false)
	case *Inverter:
		tv.visualizeNode(builder, n.Child, newPrefix, true, false)
	case *Repeater:
		limit := "∞"
		if !n.infinite {
			limit = fmt.Sprintf("%d", n.Count)
		}
		builder.WriteString(newPrefix)
		builder.WriteString(fmt.Sprintf("Repeat count: %s\n", limit))
		tv.visualizeNode(builder, n.Child, newPrefix, true, false)
	case *UntilSuccess:
		tv.visualizeNode(builder, n.Child, newPrefix, true, false)
	case *UntilFailure:
		tv.visualizeNode(builder, n.Child, newPrefix, true, false)
	}
}

// getNodeName returns a human-readable name for a behavior tree node.
func (tv *TreeVisualizer) getNodeName(node Behavior) string {
	// Check if node implements NodeVisualizer
	if visualizer, ok := node.(NodeVisualizer); ok {
		return visualizer.VisualizeNode()
	}

	// Get type name without package prefix
	typeName := reflect.TypeOf(node).String()
	if strings.Contains(typeName, ".") {
		typeName = typeName[strings.LastIndex(typeName, ".")+1:]
	}

	// Remove pointer operator if present
	typeName = strings.TrimPrefix(typeName, "*")

	// Handle special cases
	switch n := node.(type) {
	case *Action:
		return "Action"
	case *Condition:
		return "Condition"
	case *Conditional:
		return "Conditional"
	case *Sequence:
		return "Sequence"
	case *Selector:
		return "Selector"
	case *PrioritySelector:
		return "PrioritySelector"
	case *Parallel:
		policy := ""
		switch n.policy {
		case RequireOne:
			policy = "RequireOne"
		case RequireAll:
			policy = "RequireAll"
		case SuccessOnAll:
			policy = "SuccessOnAll"
		case SuccessOnOne:
			policy = "SuccessOnOne"
		}
		return fmt.Sprintf("Parallel(%s)", policy)
	case *Inverter:
		return "Inverter"
	case *Repeater:
		return "Repeater"
	case *UntilSuccess:
		return "UntilSuccess"
	case *UntilFailure:
		return "UntilFailure"
	default:
		return typeName
	}
}

// SaveTreeSnapshot captures the current state of a behavior tree during execution.
type SaveTreeSnapshot struct {
	statusMap map[Behavior]RunStatus
}

// NewSaveTreeSnapshot creates a new tree snapshot decorator.
func NewSaveTreeSnapshot() *SaveTreeSnapshot {
	return &SaveTreeSnapshot{
		statusMap: make(map[Behavior]RunStatus),
	}
}

// Visualize returns a visualization of the tree with node statuses.
func (s *SaveTreeSnapshot) Visualize(root Behavior) string {
	return NewTreeVisualizer(root).WithNodeStatuses(s.statusMap).Visualize()
}

// GetStatusMap returns the collected node statuses.
func (s *SaveTreeSnapshot) GetStatusMap() map[Behavior]RunStatus {
	return s.statusMap
}

// Tick records the status of the node and propagates the tick to the child.
func (s *SaveTreeSnapshot) Tick(ctx BehaviorContext, node Behavior) RunStatus {
	status := node.Tick(ctx)
	s.statusMap[node] = status
	return status
}
