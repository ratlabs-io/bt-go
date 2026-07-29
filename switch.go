package bt

// KeyFunc derives a case key from the context for Switch selection.
type KeyFunc func(env Env) string

// Switch selects a child by a dynamic string key.
//
// KeyFunc is evaluated every Tick. If Cases[key] exists it is ticked;
// otherwise Default is ticked. If neither matches, Switch returns Failure.
//
// When the selected case changes while the previous case was Running, the
// abandoned case is Halted. Halt propagates to the last Running case.
type Switch struct {
	KeyFunc KeyFunc
	Cases   map[string]Behavior
	Default Behavior
	// lastCase is the Behavior that returned Running on a prior tick (for Halt).
	lastCase Behavior
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
	var chosen Behavior
	if s.KeyFunc == nil {
		chosen = s.Default
	} else {
		key := s.KeyFunc(env)
		if behavior, ok := s.Cases[key]; ok && behavior != nil {
			chosen = behavior
		} else {
			chosen = s.Default
		}
	}

	if chosen == nil {
		s.abortCase(env, nil)
		return Failure
	}

	status := chosen.Tick(env)

	if s.lastCase != nil && s.lastCase != chosen {
		s.abortCase(env, chosen)
	}

	if status == Running {
		s.lastCase = chosen
	} else {
		s.lastCase = nil
	}
	return status
}

// Halt aborts the last running case.
func (s *Switch) Halt(env Env) {
	s.abortCase(env, nil)
}

func (s *Switch) abortCase(env Env, keep Behavior) {
	if s.lastCase != nil && s.lastCase != keep {
		Halt(env, s.lastCase)
	}
	if keep == nil {
		s.lastCase = nil
	}
}
