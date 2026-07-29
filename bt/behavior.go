// Package bt is a composable behavior tree library for Go.
//
// A behavior tree is a hierarchy of Behavior nodes. Each tick, a node returns
// one of Success, Failure, or Running. Leaf nodes (Action, Condition) do work;
// composites (Sequence, Selector, Parallel, …) combine children; decorators
// (Inverter, Repeater, …) wrap a single child.
//
// # Environment (Env)
//
// Nodes receive an Env on every Tick. Env is not a context.Context:
//
//   - env.Context()  — stdlib context for cancellation/deadlines only
//   - env.Blackboard() / Set/Get — mutable agent/world state
//
// Put agent data on the blackboard, never in context.WithValue.
// Trees are typically ticked by a TreeRunner or by calling Behavior.Tick directly.
package bt

// RunStatus is the result of ticking a behavior node.
type RunStatus int

const (
	// Success means the node completed its task successfully.
	Success RunStatus = iota
	// Failure means the node failed to complete its task.
	Failure
	// Running means the node is still in progress and should be ticked again.
	Running
)

// String returns a human-readable name for the status.
func (rs RunStatus) String() string {
	switch rs {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Running:
		return "Running"
	default:
		return "Unknown"
	}
}

// Behavior is implemented by every node in a behavior tree.
type Behavior interface {
	// Tick executes one step of the node and returns its status.
	Tick(env Env) RunStatus
}

// Composite is embedded by multi-child control nodes.
// Children are public so callers and tools (e.g. visualizers) can inspect the tree.
type Composite struct {
	Children []Behavior
}

// GetChildren returns the composite's child nodes.
func (c *Composite) GetChildren() []Behavior {
	return c.Children
}

// ChildrenProvider is implemented by nodes that expose multiple children.
// Used by tooling (visualization) without type-switching every composite.
type ChildrenProvider interface {
	GetChildren() []Behavior
}
