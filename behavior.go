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
)

// ToString method for RunStatus
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

// Behavior defines the interface for all behavior tree nodes.
type Behavior interface {
	Tick(ctx BehaviorContext) RunStatus
}

// Composite is the base type for nodes that have children in a behavior tree.
type Composite struct {
	Children []Behavior
}
