package bt

import (
	"fmt"
	"reflect"
	"strings"
)

// NodeVisualizer lets custom nodes supply their own label for tree dumps.
type NodeVisualizer interface {
	VisualizeNode() string
}

// TreeVisualizer renders a behavior tree as indented text.
type TreeVisualizer struct {
	root           Behavior
	showNodeStatus bool
	nodeStatuses   map[Behavior]RunStatus
}

// NewTreeVisualizer creates a visualizer for root.
func NewTreeVisualizer(root Behavior) *TreeVisualizer {
	return &TreeVisualizer{
		root:         root,
		nodeStatuses: make(map[Behavior]RunStatus),
	}
}

// WithNodeStatuses enables status annotations from the given map.
func (tv *TreeVisualizer) WithNodeStatuses(statuses map[Behavior]RunStatus) *TreeVisualizer {
	tv.showNodeStatus = true
	tv.nodeStatuses = statuses
	return tv
}

// Visualize returns a multi-line text representation of the tree.
func (tv *TreeVisualizer) Visualize() string {
	var builder strings.Builder
	tv.visualizeNode(&builder, tv.root, "", true, true)
	return builder.String()
}

func (tv *TreeVisualizer) visualizeNode(builder *strings.Builder, node Behavior, prefix string, isLast bool, isRoot bool) {
	if node == nil {
		return
	}

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

	nodeName := tv.getNodeName(node)
	if tv.showNodeStatus {
		if status, ok := tv.nodeStatuses[node]; ok {
			nodeName = fmt.Sprintf("%s [%s]", nodeName, status)
		}
	}
	builder.WriteString(nodeName)
	builder.WriteString("\n")

	// Special shapes first (Repeater needs an extra annotation line before Decorator).
	switch n := node.(type) {
	case *Repeater:
		limit := "∞"
		if !n.infinite {
			limit = fmt.Sprintf("%d", n.Count)
		}
		builder.WriteString(newPrefix)
		builder.WriteString(fmt.Sprintf("Repeat count: %s\n", limit))
		tv.visualizeNode(builder, n.Child, newPrefix, true, false)
		return
	case *Conditional:
		tv.visualizeNode(builder, n.Condition, newPrefix, false, false)
		tv.visualizeNode(builder, n.Action, newPrefix, true, false)
		return
	case *BinarySelector:
		tv.visualizeNode(builder, n.Condition, newPrefix, false, false)
		tv.visualizeNode(builder, n.IfTrue, newPrefix, false, false)
		tv.visualizeNode(builder, n.IfFalse, newPrefix, true, false)
		return
	case *Switch:
		// Map iteration order is randomized; list cases then default.
		i := 0
		total := len(n.Cases)
		if n.Default != nil {
			total++
		}
		for key, child := range n.Cases {
			lastCase := i == total-1
			caseConnector := "├── "
			casePrefix := newPrefix + "│   "
			if lastCase {
				caseConnector = "└── "
				casePrefix = newPrefix + "    "
			}
			builder.WriteString(newPrefix)
			builder.WriteString(caseConnector)
			builder.WriteString(fmt.Sprintf("case %q\n", key))
			tv.visualizeNode(builder, child, casePrefix, true, false)
			i++
		}
		if n.Default != nil {
			builder.WriteString(newPrefix)
			builder.WriteString("└── default\n")
			tv.visualizeNode(builder, n.Default, newPrefix+"    ", true, false)
		}
		return
	}

	if d, ok := node.(Decorator); ok {
		tv.visualizeNode(builder, d.GetChild(), newPrefix, true, false)
		return
	}

	if cp, ok := node.(ChildrenProvider); ok {
		children := cp.GetChildren()
		for i, child := range children {
			tv.visualizeNode(builder, child, newPrefix, i == len(children)-1, false)
		}
	}
}

func (tv *TreeVisualizer) getNodeName(node Behavior) string {
	if visualizer, ok := node.(NodeVisualizer); ok {
		return visualizer.VisualizeNode()
	}

	typeName := reflect.TypeOf(node).String()
	if strings.Contains(typeName, ".") {
		typeName = typeName[strings.LastIndex(typeName, ".")+1:]
	}
	typeName = strings.TrimPrefix(typeName, "*")

	switch n := node.(type) {
	case *Action:
		return "Action"
	case *Condition:
		return "Condition"
	case *Conditional:
		return "Conditional"
	case *Sequence:
		return "Sequence"
	case *MemorySequence:
		return "MemorySequence"
	case *Selector:
		return "Selector"
	case *MemorySelector:
		return "MemorySelector"
	case *Parallel:
		return fmt.Sprintf("Parallel(%s)", n.Policy())
	case *BinarySelector:
		return "BinarySelector"
	case *Switch:
		return "Switch"
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

// StatusRecorder records the status of nodes as they are ticked.
// It is a debugging aid, not a tree node itself.
//
// Prefer wrapping ticks you care about:
//
//	rec := bt.NewStatusRecorder()
//	status := rec.Tick(ctx, root) // records root only
//
// For full-tree status maps, wrap individual leaves or use a custom decorator.
type StatusRecorder struct {
	statusMap map[Behavior]RunStatus
}

// NewStatusRecorder creates an empty status recorder.
func NewStatusRecorder() *StatusRecorder {
	return &StatusRecorder{statusMap: make(map[Behavior]RunStatus)}
}

// NewSaveTreeSnapshot is a deprecated alias for NewStatusRecorder.
//
// Deprecated: use NewStatusRecorder.
func NewSaveTreeSnapshot() *StatusRecorder {
	return NewStatusRecorder()
}

// Visualize renders root with recorded statuses annotated.
func (s *StatusRecorder) Visualize(root Behavior) string {
	return NewTreeVisualizer(root).WithNodeStatuses(s.statusMap).Visualize()
}

// GetStatusMap returns the map of recorded node statuses.
func (s *StatusRecorder) GetStatusMap() map[Behavior]RunStatus {
	return s.statusMap
}

// Tick runs node, records its status, and returns that status.
// Only the node itself is recorded — not its descendants.
func (s *StatusRecorder) Tick(ctx BehaviorContext, node Behavior) RunStatus {
	if node == nil {
		return Failure
	}
	status := node.Tick(ctx)
	s.statusMap[node] = status
	return status
}
