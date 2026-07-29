package bt

// Instrument returns a new tree that reports every node tick to after.
//
// The original tree is not modified. Structure is preserved: composites,
// decorators, BinarySelector, Switch, and Conditional are rebuilt with
// instrumented children, then each rebuilt node is wrapped in Observing.
//
// after may be nil (structure is still cloned/wrapped). Pointer identity of
// nodes differs from root — use the returned tree for Tick and visualization.
//
// This is the deep-observation alternative to wrapping individual leaves with
// NewObserving. Prefer per-node Observing when only a few nodes matter.
func Instrument(root Behavior, after TickObserver) Behavior {
	if root == nil {
		return nil
	}
	return observe(cloneInstrumented(root, after), after)
}

// InstrumentRecorder is like Instrument but records each node status into rec.
// If rec is nil, a new StatusRecorder is created. The recorder and instrumented
// root are both returned so callers can Visualize after ticking.
func InstrumentRecorder(root Behavior, rec *StatusRecorder) (Behavior, *StatusRecorder) {
	if rec == nil {
		rec = NewStatusRecorder()
	}
	tree := Instrument(root, func(node Behavior, status RunStatus) {
		rec.statusMap[node] = status
	})
	return tree, rec
}

func observe(child Behavior, after TickObserver) Behavior {
	if child == nil {
		return nil
	}
	return NewObserving(child, after)
}

// cloneInstrumented deep-clones control structure with instrumented children.
// Leaves (Action, Condition, unknown types) are reused by pointer.
func cloneInstrumented(node Behavior, after TickObserver) Behavior {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *Sequence:
		return NewSequence(instrumentChildren(n.Children, after)...)
	case *MemorySequence:
		return NewMemorySequence(instrumentChildren(n.Children, after)...)
	case *Selector:
		return NewSelector(instrumentChildren(n.Children, after)...)
	case *MemorySelector:
		return NewMemorySelector(instrumentChildren(n.Children, after)...)
	case *Parallel:
		children := instrumentChildren(n.Children, after)
		if n.Concurrent() {
			return NewConcurrentParallel(n.Policy(), children...)
		}
		return NewParallel(n.Policy(), children...)
	case *Inverter:
		return NewInverter(observe(cloneInstrumented(n.Child, after), after))
	case *Repeater:
		return NewRepeater(observe(cloneInstrumented(n.Child, after), after), n.Count)
	case *UntilSuccess:
		return NewUntilSuccess(observe(cloneInstrumented(n.Child, after), after))
	case *UntilFailure:
		return NewUntilFailure(observe(cloneInstrumented(n.Child, after), after))
	case *Named:
		return NewNamed(n.Name, observe(cloneInstrumented(n.Child, after), after))
	case *Observing:
		// Preserve existing observer by wrapping instrumented child; outer
		// Instrument Observing still records the rebuilt node.
		inner := observe(cloneInstrumented(n.Child, after), after)
		return NewObserving(inner, n.AfterTick)
	case *AbortHook:
		return NewAbortHook(observe(cloneInstrumented(n.Child, after), after), n.OnAbort)
	case *BinarySelector:
		return NewBinarySelector(
			observe(cloneInstrumented(n.Condition, after), after),
			observe(cloneInstrumented(n.IfTrue, after), after),
			observe(cloneInstrumented(n.IfFalse, after), after),
		)
	case *Switch:
		cases := make(map[string]Behavior, len(n.Cases))
		for k, v := range n.Cases {
			cases[k] = observe(cloneInstrumented(v, after), after)
		}
		return NewSwitch(n.KeyFunc, cases, observe(cloneInstrumented(n.Default, after), after))
	case *Conditional:
		// Condition leaf is reused; Action is instrumented.
		var cond *Condition
		if n.Condition != nil {
			cond = n.Condition
		}
		return NewConditional(cond, observe(cloneInstrumented(n.Action, after), after))
	}

	// Unknown decorator/composite types: wrap as-is (cannot rebuild safely).
	// Leaves and unrecognized nodes: share pointer (status recorded via parent Observing).
	return node
}

func instrumentChildren(children []Behavior, after TickObserver) []Behavior {
	out := make([]Behavior, len(children))
	for i, c := range children {
		out[i] = observe(cloneInstrumented(c, after), after)
	}
	return out
}
