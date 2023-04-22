package bt

// Switch represents a behavior tree node that selects one of multiple child nodes based on a key.
type Switch struct {
	Key     Key
	Cases   map[Key]Behavior
	Default Behavior
	Current Behavior
}

// Tick updates the behavior's state based on the given context.
func (s *Switch) Tick(ctx *BehaviorContext) RunStatus {
	// Get the current behavior for the given key, if any
	if s.Current == nil {
		s.Current = s.Cases[s.Key]
	}
	// If the current behavior is not running, tick it
	if s.Current != nil {
		status := s.Current.Tick(ctx)
		if status == Failure {
			s.Current = nil
			return Failure
		} else if status == Success {
			s.Current = nil
			return Success
		} else {
			return Running
		}
	}
	// If the current behavior is nil, check if there is a case for the given key and set it as the current behavior
	if behavior, ok := s.Cases[s.Key]; ok {
		s.Current = behavior
		return Running
	}
	// If there is no case for the given key, use the default behavior if it is defined
	if s.Default != nil {
		return s.Default.Tick(ctx)
	}
	// If there is no case and no default, return failure
	return Failure
}
