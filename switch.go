package bt

// KeyFunc is a function type that takes a BehaviorContext and returns a string key.
type KeyFunc func(ctx BehaviorContext) string

// Switch represents a behavior tree node that selects one of multiple child nodes based on a key.
type Switch struct {
	KeyFunc KeyFunc
	Cases   map[string]Behavior
	Default Behavior
}

// NewSwitch returns a new Switch behavior with the given key function, cases, and default behavior.
func NewSwitch(keyFunc KeyFunc, cases map[string]Behavior, defaultBehavior Behavior) *Switch {
	return &Switch{
		KeyFunc: keyFunc,
		Cases:   cases,
		Default: defaultBehavior,
	}
}

// Tick updates the behavior's state based on the given BehaviorContext.
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
