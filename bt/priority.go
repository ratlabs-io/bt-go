package bt

// PrioritySelector is a type alias for Selector.
//
// Historically this library had two types with identical tick logic. Both are
// reactive selectors: children are always evaluated from the first child
// (highest priority) on every Tick. Prefer NewSelector for new code;
// NewPrioritySelector remains for API compatibility.
type PrioritySelector = Selector

// NewPrioritySelector creates a priority selector (identical to NewSelector).
func NewPrioritySelector(children ...Behavior) *PrioritySelector {
	return NewSelector(children...)
}
