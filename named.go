package bt

// Named wraps a child with a human-readable label for visualization and logs.
// Tick is a pure pass-through.
type Named struct {
	BaseDecorator
	Name string
}

// NewNamed creates a Named decorator. Empty name falls back to "Named" in dumps.
func NewNamed(name string, child Behavior) *Named {
	return &Named{
		BaseDecorator: BaseDecorator{Child: child},
		Name:          name,
	}
}

// Tick runs the child unchanged.
func (n *Named) Tick(env Env) RunStatus {
	if n.Child == nil {
		return Failure
	}
	return n.Child.Tick(env)
}

// VisualizeNode returns the custom name for TreeVisualizer.
func (n *Named) VisualizeNode() string {
	if n.Name == "" {
		return "Named"
	}
	return n.Name
}

// Halt propagates abort to the child.
func (n *Named) Halt(env Env) {
	Halt(env, n.Child)
}
