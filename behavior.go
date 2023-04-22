package bt

// RunStatus represents the possible statuses of a behavior when it is ticked.
type RunStatus int

// Key is a string that represents a behavior's unique identifier.
type Key string

const (
	// Success represents a behavior that has completed successfully.
	Success RunStatus = iota
	// Failure represents a behavior that has completed unsuccessfully.
	Failure
	// Running represents a behavior that is still in progress.
	Running
	// Invalid represents a behavior that has not been initialized.
	Invalid
	// Aborted represents a behavior that has been aborted.
	Aborted
)

// Behavior is the interface that all behavior tree nodes must implement.
type Behavior interface {
	// Tick is called to update the behavior's state. It takes a context as an argument.
	Tick(ctx *BehaviorContext) RunStatus
	// Abort is called to abort the behavior's execution. It takes a context as an argument.
	Abort(ctx *BehaviorContext)
}

// Composite is the base type for nodes that have children in a behavior tree.
type Composite struct {
	Children []Behavior
}
