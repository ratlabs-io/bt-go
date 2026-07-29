package bt

// Decorator wraps a single child and alters when it runs or what status it reports.
type Decorator interface {
	Behavior
	SetChild(child Behavior)
	GetChild() Behavior
}

// BaseDecorator holds the child pointer shared by concrete decorators.
type BaseDecorator struct {
	Child Behavior
}

// SetChild sets the decorated child.
func (d *BaseDecorator) SetChild(child Behavior) {
	d.Child = child
}

// GetChild returns the decorated child.
func (d *BaseDecorator) GetChild() Behavior {
	return d.Child
}

// Inverter swaps Success ↔ Failure. Running is unchanged.
type Inverter struct {
	BaseDecorator
}

// NewInverter creates an Inverter around child.
func NewInverter(child Behavior) *Inverter {
	return &Inverter{BaseDecorator: BaseDecorator{Child: child}}
}

// Tick runs the child and inverts terminal statuses.
func (i *Inverter) Tick(env Env) RunStatus {
	if i.Child == nil {
		return Failure
	}
	switch status := i.Child.Tick(env); status {
	case Success:
		return Failure
	case Failure:
		return Success
	default:
		return status
	}
}

// Repeater runs its child a fixed number of successful times.
// Each Tick advances at most one successful child completion.
// Failure from the child aborts and resets the counter.
// count <= 0 means infinite (always returns Running after a successful child tick).
type Repeater struct {
	BaseDecorator
	Count    int
	counter  int
	infinite bool
}

// NewRepeater creates a Repeater. count <= 0 means repeat forever.
func NewRepeater(child Behavior, count int) *Repeater {
	return &Repeater{
		BaseDecorator: BaseDecorator{Child: child},
		Count:         count,
		infinite:      count <= 0,
	}
}

// Tick executes the child and tracks successful completions toward Count.
func (r *Repeater) Tick(env Env) RunStatus {
	if r.Child == nil {
		return Failure
	}

	if !r.infinite && r.counter >= r.Count {
		r.counter = 0
		return Success
	}

	status := r.Child.Tick(env)
	if status == Running {
		return Running
	}
	if status == Failure {
		r.counter = 0
		return Failure
	}

	r.counter++
	if r.infinite || r.counter < r.Count {
		return Running
	}
	r.counter = 0
	return Success
}

// UntilSuccess ticks the child until it returns Success.
// Failure and Running both yield Running from the decorator.
type UntilSuccess struct {
	BaseDecorator
}

// NewUntilSuccess creates an UntilSuccess decorator.
func NewUntilSuccess(child Behavior) *UntilSuccess {
	return &UntilSuccess{BaseDecorator: BaseDecorator{Child: child}}
}

// Tick returns Success only when the child succeeds; otherwise Running.
func (u *UntilSuccess) Tick(env Env) RunStatus {
	if u.Child == nil {
		return Failure
	}
	if u.Child.Tick(env) == Success {
		return Success
	}
	return Running
}

// UntilFailure ticks the child until it returns Failure.
// When the child fails, the decorator returns Success (the wait succeeded).
type UntilFailure struct {
	BaseDecorator
}

// NewUntilFailure creates an UntilFailure decorator.
func NewUntilFailure(child Behavior) *UntilFailure {
	return &UntilFailure{BaseDecorator: BaseDecorator{Child: child}}
}

// Tick returns Success when the child fails; otherwise Running.
func (u *UntilFailure) Tick(env Env) RunStatus {
	if u.Child == nil {
		return Failure
	}
	if u.Child.Tick(env) == Failure {
		return Success
	}
	return Running
}
