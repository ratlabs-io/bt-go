package bt

// RunStatus represents the possible execution states of a behavior node when it is ticked during the behavior tree traversal.
type RunStatus int

// Key is a string type alias that represents a behavior's unique identifier within the tree.
type Key string

const (
	// Success indicates that a behavior node has completed its task successfully.
	Success RunStatus = iota
	// Failure indicates that a behavior node has encountered an error or failed to complete its task.
	Failure
	// Running indicates that a behavior node is still in progress and has not yet completed.
	Running
)

// String converts a RunStatus value to its string representation for easier debugging and logging.
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

// Behavior defines the interface that all behavior tree nodes must implement.
// It provides a method to execute the node's logic within the context of the tree.
type Behavior interface {
	// Tick executes the behavior node's logic using the provided context and returns its current status.
	Tick(ctx BehaviorContext) RunStatus
}

// Composite is a base struct for behavior nodes that can have child nodes.
// It is used to build composite behaviors like sequences, selectors, etc.
type Composite struct {
	Children []Behavior // Children holds the list of child behavior nodes.
}
