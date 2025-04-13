package bt

// KeyFunc is a function type that takes a BehaviorContext and returns a string key used for selecting a behavior in a Switch node.
type KeyFunc func(ctx BehaviorContext) string

// Switch represents a control node in a behavior tree that selects one of multiple child behaviors to execute based on a key.
// The key is determined by a provided function, and behaviors are mapped to specific keys with a default behavior as a fallback.
type Switch struct {
	KeyFunc KeyFunc             // KeyFunc is the function used to determine the selection key from the context.
	Cases   map[string]Behavior // Cases maps keys to specific behaviors to be executed.
	Default Behavior            // Default is the fallback behavior if no case matches the key.
}

// NewSwitch creates a new Switch behavior with the given key function, cases, and default behavior.
// The key function determines which behavior to execute based on the context, selecting from the provided cases or falling back to the default.
func NewSwitch(keyFunc KeyFunc, cases map[string]Behavior, defaultBehavior Behavior) *Switch {
	return &Switch{
		KeyFunc: keyFunc,
		Cases:   cases,
		Default: defaultBehavior,
	}
}

// Tick executes the Switch node's logic using the given BehaviorContext.
// It determines the key from the context, selects the corresponding behavior from the cases, or uses the default behavior if no match is found.
// If no behavior is selected and no default is provided, it returns Failure.
func (s *Switch) Tick(ctx BehaviorContext) RunStatus {
	key := s.KeyFunc(ctx)
	if behavior, ok := s.Cases[key]; ok {
		return behavior.Tick(ctx)
	}
	if s.Default != nil {
		return s.Default.Tick(ctx)
	}
	return Failure
}
