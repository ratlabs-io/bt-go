package bt

// Decorator is a base interface for all decorator nodes in a behavior tree.
// Decorators modify the behavior of their child node by either conditioning
// when it runs or transforming its return status.
type Decorator interface {
	Behavior
	SetChild(child Behavior)
	GetChild() Behavior
}

// BaseDecorator provides a common implementation for decorator nodes.
type BaseDecorator struct {
	Child Behavior
}

// SetChild sets the child node for this decorator.
func (d *BaseDecorator) SetChild(child Behavior) {
	d.Child = child
}

// GetChild returns the child node of this decorator.
func (d *BaseDecorator) GetChild() Behavior {
	return d.Child
}

// Inverter is a decorator that inverts the result of its child node:
// Success becomes Failure and vice versa. Running remains unchanged.
type Inverter struct {
	BaseDecorator
}

// NewInverter creates a new Inverter with the given child node.
func NewInverter(child Behavior) *Inverter {
	return &Inverter{
		BaseDecorator: BaseDecorator{Child: child},
	}
}

// Tick executes the child node and inverts its result.
func (i *Inverter) Tick(ctx BehaviorContext) RunStatus {
	if i.Child == nil {
		return Failure
	}

	status := i.Child.Tick(ctx)

	switch status {
	case Success:
		return Failure
	case Failure:
		return Success
	default:
		return status
	}
}

// Repeater is a decorator that repeats its child node a specified number of times.
type Repeater struct {
	BaseDecorator
	Count    int
	counter  int
	infinite bool
}

// NewRepeater creates a new Repeater with the given child node and count.
// If count is <= 0, the repeater will run indefinitely.
func NewRepeater(child Behavior, count int) *Repeater {
	infinite := count <= 0
	return &Repeater{
		BaseDecorator: BaseDecorator{Child: child},
		Count:         count,
		counter:       0,
		infinite:      infinite,
	}
}

// Tick executes the child node repeatedly up to the specified count.
func (r *Repeater) Tick(ctx BehaviorContext) RunStatus {
	if r.Child == nil {
		return Failure
	}

	if !r.infinite && r.counter >= r.Count {
		r.counter = 0
		return Success
	}

	status := r.Child.Tick(ctx)

	if status == Running {
		return Running
	}

	if status == Failure {
		r.counter = 0
		return Failure
	}

	// Child succeeded
	r.counter++
	if r.infinite || r.counter < r.Count {
		return Running
	}

	r.counter = 0
	return Success
}

// UntilSuccess repeats the child node until it succeeds.
type UntilSuccess struct {
	BaseDecorator
}

// NewUntilSuccess creates a new UntilSuccess decorator with the given child.
func NewUntilSuccess(child Behavior) *UntilSuccess {
	return &UntilSuccess{
		BaseDecorator: BaseDecorator{Child: child},
	}
}

// Tick executes the child node until it succeeds.
func (u *UntilSuccess) Tick(ctx BehaviorContext) RunStatus {
	if u.Child == nil {
		return Failure
	}

	status := u.Child.Tick(ctx)

	if status == Success {
		return Success
	}

	return Running
}

// UntilFailure repeats the child node until it fails.
type UntilFailure struct {
	BaseDecorator
}

// NewUntilFailure creates a new UntilFailure decorator with the given child.
func NewUntilFailure(child Behavior) *UntilFailure {
	return &UntilFailure{
		BaseDecorator: BaseDecorator{Child: child},
	}
}

// Tick executes the child node until it fails.
func (u *UntilFailure) Tick(ctx BehaviorContext) RunStatus {
	if u.Child == nil {
		return Failure
	}

	status := u.Child.Tick(ctx)

	if status == Failure {
		return Success
	}

	return Running
}
