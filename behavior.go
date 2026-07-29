// Package bt is a composable behavior tree library for Go.
//
// Import: "github.com/ratlabs-io/bt-go" (package name bt).
//
// A behavior tree is a hierarchy of Behavior nodes. Each tick, a node returns
// one of Success, Failure, or Running. Leaf nodes (Action, Condition) do work;
// composites (Sequence, Selector, Parallel, …) combine children; decorators
// (Inverter, Repeater, Named, Observing, AbortHook, …) wrap a single child.
//
// # Environment (Env)
//
// Nodes receive an Env on every Tick. Env is not a context.Context:
//
//   - env.Context() — stdlib context for cancellation/deadlines only
//   - blackboard via Set/Get/GetAs — mutable agent/world state
//
// Put agent data on the blackboard, never in context.WithValue.
//
// # Abort (Halt)
//
// When a parent abandons a Running child (preemption, branch switch, cancel),
// Halt is called on Haltable nodes so they can clean up (see AbortHook).
// Control-flow nodes that can abandon work (Sequence, Selector, memory
// composites, Parallel, BinarySelector, Switch, Conditional, decorators)
// implement Haltable and forward Halt to the abandoned child.
//
// # Observation
//
// NewObserving wraps a single node. Instrument / InstrumentRecorder wrap an
// entire tree so every tick reports to a callback or StatusRecorder.
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
