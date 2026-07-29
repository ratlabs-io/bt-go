package bt

// TickObserver is called after a watched node finishes a Tick.
type TickObserver func(node Behavior, status RunStatus)

// Observing wraps a child and reports each tick result to AfterTick.
// Only this node is observed — wrap additional nodes to observe deeper.
type Observing struct {
	BaseDecorator
	// AfterTick is invoked after the child is ticked (including when child is nil → Failure).
	AfterTick TickObserver
}

// NewObserving creates an Observing decorator. after may be nil (no-op).
func NewObserving(child Behavior, after TickObserver) *Observing {
	return &Observing{
		BaseDecorator: BaseDecorator{Child: child},
		AfterTick:     after,
	}
}

// Tick runs the child and notifies AfterTick.
func (o *Observing) Tick(env Env) RunStatus {
	var status RunStatus = Failure
	if o.Child != nil {
		status = o.Child.Tick(env)
	}
	if o.AfterTick != nil {
		// Report the child when present; otherwise the Observing node itself.
		node := Behavior(o)
		if o.Child != nil {
			node = o.Child
		}
		o.AfterTick(node, status)
	}
	return status
}

// Halt propagates abort to the child.
func (o *Observing) Halt(env Env) {
	Halt(env, o.Child)
}

// AbortHook runs a callback when the node is Halted while still considered active
// (last tick returned Running). Use it for cleanup when a reactive parent preempts
// a long-running branch (stop pathing, clear target, etc.).
type AbortHook struct {
	BaseDecorator
	OnAbort func(env Env)
	active  bool
}

// NewAbortHook wraps child with an abort callback.
func NewAbortHook(child Behavior, onAbort func(env Env)) *AbortHook {
	return &AbortHook{
		BaseDecorator: BaseDecorator{Child: child},
		OnAbort:       onAbort,
	}
}

// Tick runs the child and tracks whether it is mid-run.
func (a *AbortHook) Tick(env Env) RunStatus {
	if a.Child == nil {
		a.active = false
		return Failure
	}
	status := a.Child.Tick(env)
	a.active = status == Running
	return status
}

// Halt invokes OnAbort if the child was Running, then Halt the child.
func (a *AbortHook) Halt(env Env) {
	if a.active {
		if a.OnAbort != nil {
			a.OnAbort(env)
		}
		a.active = false
	}
	Halt(env, a.Child)
}
