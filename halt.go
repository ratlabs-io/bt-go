package bt

// Haltable is implemented by nodes that need cleanup when a parent aborts them
// without delivering a terminal Success/Failure tick (preemption, cancel, Reset).
//
// Halt must be safe to call when the node is not running (no-op).
// Composites that track a running child should Halt that child and clear memory.
type Haltable interface {
	Halt(env Env)
}

// Halt calls Halt on node if it implements Haltable; otherwise it is a no-op.
func Halt(env Env, node Behavior) {
	if node == nil {
		return
	}
	if h, ok := node.(Haltable); ok {
		h.Halt(env)
	}
}

// HaltAll calls Halt on each node.
func HaltAll(env Env, nodes ...Behavior) {
	for _, n := range nodes {
		Halt(env, n)
	}
}
