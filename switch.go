package bt

// KeyFunc derives a case key from the context for Switch selection.
type KeyFunc func(env Env) string

// Switch selects a child by a dynamic string key.
//
// KeyFunc is evaluated every Tick. If Cases[key] exists it is ticked;
// otherwise Default is ticked. If neither matches, Switch returns Failure.
type Switch struct {
	KeyFunc KeyFunc
	Cases   map[string]Behavior
	Default Behavior
}

// NewSwitch creates a Switch. cases may be nil (only Default will run).
func NewSwitch(keyFunc KeyFunc, cases map[string]Behavior, defaultBehavior Behavior) *Switch {
	if cases == nil {
		cases = map[string]Behavior{}
	}
	return &Switch{
		KeyFunc: keyFunc,
		Cases:   cases,
		Default: defaultBehavior,
	}
}

// Tick selects and runs the matching case or default.
func (s *Switch) Tick(env Env) RunStatus {
	if s.KeyFunc == nil {
		if s.Default != nil {
			return s.Default.Tick(env)
		}
		return Failure
	}
	key := s.KeyFunc(env)
	if behavior, ok := s.Cases[key]; ok && behavior != nil {
		return behavior.Tick(env)
	}
	if s.Default != nil {
		return s.Default.Tick(env)
	}
	return Failure
}
