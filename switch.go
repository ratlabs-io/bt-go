package bt

// Switch represents a behavior tree node that selects one of multiple child nodes based on a key.
type Switch struct {
	Key     Key
	Cases   map[Key]Behavior
	Default Behavior
}

// Tick updates the behavior's state based on the given context.
func (s *Switch) Tick(ctx *BehaviorContext) RunStatus {
	if behavior, ok := s.Cases[s.Key]; ok {
		return behavior.Tick(ctx)
	}
	if s.Default != nil {
		return s.Default.Tick(ctx)
	}
	return Failure
}
